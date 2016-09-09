require 'fileutils'

describe 'pre-commit' do
  let(:tmp_dir)  { 'spec/tmp' }
  let(:hook_dir) { "#{tmp_dir}/.quickhook/pre-commit" }

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
    Dir.chdir(tmp_dir) do
      output = `../../quickhook hook pre-commit --no-color`

      Result.new $?.exitstatus, output
    end
  end

  before do
    FileUtils.mkdir_p tmp_dir
    FileUtils.mkdir_p hook_dir

    Dir.chdir(tmp_dir) do
      system 'git init --quiet .'
      system 'echo "Changed!" > example'
      system 'git add example'
    end
  end

  after do
    FileUtils.rm_r [
      'spec/tmp/.git',
      'spec/tmp/.quickhook',
      'spec/tmp/example',
    ]
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
end
