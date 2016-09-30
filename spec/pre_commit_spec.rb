require 'fileutils'
require 'pty'

describe 'pre-commit' do
  let(:hook_dir) { '.quickhook/pre-commit' }

  def system(command)
    Kernel.system(command) || exit($?.exitstatus)
  end

  def write_file(path, contents)
    File.open(path, 'w') do |f|
      f.write contents
    end
  end

  Result = Struct.new(:status, :output) do
    def lines
      self.output.strip.split(/[\r\n]+/)
    end
  end

  def run_hook(pty: true, options: nil)
    command = "../../quickhook hook pre-commit #{options}"

    if pty
      run_hook_with_pty command
    else
      run_hook_without_pty command
    end
  end

  def run_hook_without_pty(command)
    output = IO.popen(command, &:read)
    Result.new $?.exitstatus, output
  end

  def run_hook_with_pty(command)
    $stdout.sync = true
    PTY.spawn command do |ttyout, ttyin, pid|
      ttyin.close

      output = ''
      begin
        loop do
          begin
            output << ttyout.read_nonblock(4096)
          rescue IO::WaitReadable
            IO.select([ttyout], nil, nil, 1)
            retry
          end
        end
      rescue Errno::EIO
      rescue EOFError
      end

      _, status = Process.waitpid2(pid)
      return Result.new(status.exitstatus, output)
    end
  end

  before do
    FileUtils.mkdir_p hook_dir

    system 'git init --quiet .'
    system 'echo "[user] \n name = example \n email = example@example.com" >> .git/config'

    system 'echo "Changed!" > example'
    system 'git add example'
  end

  after do
    FileUtils.rm_r [
      '.git',
      '.quickhook',
      'example',
      'other-example',
    ], force: true
  end

  # NOTE: `around` wraps both `before` and `after` hooks
  around do |example|
    Dir.chdir('spec/tmp') do
      example.run
    end
  end

  it "fails if any of the hooks failed" do
    write_file "#{hook_dir}/fails", "#!/bin/bash \n echo \"failed\" \n exit 1"
    system "chmod +x #{hook_dir}/*"

    result = run_hook(pty: false)
    expect(result.status).not_to eq 0
    expect(result.lines).to eq(['fails: fail', 'failed'])

    result = run_hook(pty: true, options: '--no-color')
    expect(result.lines).to eq(["fails: fail", "failed"])

    result = run_hook(pty: true)
    expect(result.lines).to eq(["fails: \e[31mfail\e[0m", "\e[31mfailed", "\e[0m"])
  end

  it "passes if all hooks pass" do
    write_file "#{hook_dir}/passes1", "#!/bin/bash \n echo \"passed\" \n exit 0"
    write_file "#{hook_dir}/passes2", "#!/bin/bash \n echo \"passed\" \n exit 0"
    system "chmod +x #{hook_dir}/*"

    result = run_hook(pty: false)
    expect(result.status).to eq 0
    expect(result.lines.sort).to eq(["passes1: ok", "passes2: ok"])

    result = run_hook(pty: true, options: '--no-color')
    expect(result.lines.sort).to eq(["passes1: ok", "passes2: ok"])

    result = run_hook(pty: true)
    expect(result.lines.sort).to eq(["passes1: \e[32mok\e[0m", "passes2: \e[32mok\e[0m"])
  end

  it 'handles deleted files' do
    system 'git commit --message "Commit example" --quiet --no-verify'

    system 'git rm example --quiet'

    system 'echo "Also changed!" > other-example'
    system 'git add other-example'

    result = run_hook

    expect(result.status).to eq 0
  end
end
