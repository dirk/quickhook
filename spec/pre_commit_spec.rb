require 'fileutils'

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
      self.output.strip.split("\n")
    end
  end

  def run_hook
    output = `../../quickhook hook pre-commit --no-color`

    Result.new $?.exitstatus, output
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

    result = run_hook

    expect(result.status).not_to eq 0
    expect(result.lines).to eq([
      'fails: fail',
      'failed',
    ])
  end

  it "passes if all hooks pass" do
    write_file "#{hook_dir}/passes1", "#!/bin/bash \n echo \"passed\" \n exit 0"
    write_file "#{hook_dir}/passes2", "#!/bin/bash \n echo \"passed\" \n exit 0"
    system "chmod +x #{hook_dir}/*"

    result = run_hook

    expect(result.status).to eq 0
    expect(result.lines.sort).to eq([
      'passes1: ok',
      'passes2: ok',
    ])
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
