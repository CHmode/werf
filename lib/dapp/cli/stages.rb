require 'mixlib/cli'

module Dapp
  class CLI
    # CLI stages subcommand
    class Stages < CLI
      SUBCOMMANDS = %w(flush cleanup).freeze

      banner <<BANNER.freeze
Version: #{Dapp::VERSION}

Available subcommands: (for details, dapp SUB-COMMAND --help)

dapp stages flush
dapp stages cleanup

Options:
BANNER
    end
  end
end
