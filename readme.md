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

# Documentation

[Here.](https://github.com/insomnimus/inscript/wiki)

# TODO

Examples and more documentation coming soon.

Meanwhile, though, here's a totally realistic inscript code (placeholder for more thought out examples).

```
#!/home/insomnia/go/bin/inscript

# the ':' prefix here is a shorthand for 'sync:=true'.
# redirect the output to 'echoed.txt'.
@ :echo "starting!" {
	stdout:= echoed.txt
}

# periodically read the last line from echoed.txt.
@ tail -1 echoed.txt {
	stdout:= !stdout # the whole scripts standard output
	every:= 30s # do it every 30 seconds
	times:= 4 # do this 4 times
}

# sleep a bit so there are no clashes.
# the ':' is important because commands are asynchronous by default
:sleep 15

# now lets append to echoed.txt every 30 seconds
@ date +"%H:%M:%S" {
	stdout:= echoed.txt
	every:= 30s
	times:= 3
	sync:= true # wait for the execution
}

# give tail enough time to complete its last iteration
:sleep 16

# the '!' prefix is the shorthand for
#	stdout:= !stdout
#	stderr:= !stderr
#
# the escape sequences are parsed, if double quoted.
# note that we don't use 'echo -e', thats because inscript already parses double quoted strings.
!:echo "\tI'm done!"
```

output:

```
$ ./time.ins
starting ./time.ins
starting!
"03:16:31"
"03:17:01"
"03:17:31"
        I'm done!
done ./time.ins
```
