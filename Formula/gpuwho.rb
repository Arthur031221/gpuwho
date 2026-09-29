class Gpuwho < Formula
  desc "Live per-process GPU and Neural Engine attribution for Apple Silicon"
  homepage "https://github.com/Arthur031221/gpuwho"
  url "https://github.com/Arthur031221/gpuwho/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_SHA256_AFTER_TAGGING_v0.1.0"
  license "MIT"
  head "https://github.com/Arthur031221/gpuwho.git", branch: "main"

  depends_on "go" => :build
  depends_on :macos

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "./cmd/gpuwho"
  end

  test do
    assert_match "gpuwho", shell_output("#{bin}/gpuwho --version")
    system bin/"gpuwho", "--once"
  end
end
