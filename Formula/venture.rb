# Homebrew formula for Venture
# Installation: brew install opd-ai/tap/venture
# Or: brew tap opd-ai/tap && brew install venture

class Venture < Formula
  desc "Procedural multiplayer action-RPG with 100% runtime-generated content"
  homepage "https://github.com/opd-ai/venture"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/opd-ai/venture/releases/download/v1.0.0/venture-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  depends_on "mesa" => :optional if OS.linux?

  def install
    bin.install "venture-server"
    bin.install "venture-client"
  end

  def caveats
    <<~EOS
      Venture has been installed!

      To start the game client:
        venture-client

      To start a dedicated server:
        venture-server --port 7777 --max-players 10

      For more options:
        venture-server --help
        venture-client --help

      Documentation: https://github.com/opd-ai/venture#readme
    EOS
  end

  service do
    run [opt_bin/"venture-server", "--port", "7777", "--max-players", "10"]
    keep_alive true
    working_dir var/"venture"
    log_path var/"log/venture-server.log"
    error_log_path var/"log/venture-server.error.log"
  end

  test do
    assert_match "Venture", shell_output("#{bin}/venture-server --version")
    assert_match "Venture", shell_output("#{bin}/venture-client --version")
  end
end
