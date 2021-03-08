require "open3"

workers 2
bind 'tcp://0.0.0.0:8080'
threads 1,2

app do |env|
  buf = ""

    [
      "which bundler",
      "bundle version",
      "which ruby",
      "ruby --version"
    ].each do |command|
      output, _ = Open3.capture2e(command)
      buf += "$ #{command}\n#{output}\n"
    end

  [
    200,
    {
      'Content-Type' => 'text/plain',
      'Content-Length' => buf.length.to_s
    },
    [buf]
  ]
end
