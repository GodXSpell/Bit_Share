class Bitshare < Formula
  desc "High-speed P2P mesh network file sharing system"
  homepage "https://github.com/yourusername/bitshare"
  url "https://github.com/yourusername/bitshare/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"
  head "https://github.com/yourusername/bitshare.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match "BitShare P2P Mesh Network", shell_output("#{bin}/bitshare --help")
  end
end
