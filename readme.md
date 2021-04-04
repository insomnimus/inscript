# inscript

inscript is a domain specific language for scheduling jobs concurrently.

# What Does It Do?

inscript parses an inscript file (.ins) and executes the commands within.
The main use case for inscript is for concurrently running multiple jobs with extra configuration options for each command, so it is sort of like cron and make.

# Why did you make this? What's the point?

I found myself having to write a lot of code to do some specific automation too often, inscript is my solution for this.

# Installation

You need go 1.16 or newer to compile inscript.

	go get -u -v github.com/insomnimus/inscript

or you can clone it:

	git clone https://www.github.com/insomnimus/inscript
	cd inscript
	# if you have gnu make
	make all
	# or just
	# go install

after the steps above, inscript binary should be in $GOPATH/bin (or $GOBIN if you set it manually). Make sure it's under your $PATH.

# TODO

Examples and documentation coming soon.

Meanwhile, though, here's a totally realistic inscript code (placeholder for more thought out examples).

