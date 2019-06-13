.PHONY: demo

default: help;

demo:
	go run cmd/humanize/demo.go

help:
	# humanize: a human-eye-friendly log formatter for Logrus
	#
	# Note that humanize is not intended for command-line use; it's a plugin for:
	# http://github.com/sirupsen/logrus
	#
	# A `demo` command is included for, as you might expect, demonstration purposes.
	#
	# demo: Runs the demo binary, showing what formatting will look like in your terminal
	# test: Run the tests
