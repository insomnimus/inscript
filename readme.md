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

```
#!/home/insomnia/go/bin/inscript

# the ':' before the command is to indicate that it's not async (should wait for it to exit before moving to the next command)
@ :echo hello {
	stdout:= echoed.txt # write to echoed.txt
}

# lets read from echoed.txt
@ tail -1 echoed.txt {
sync:= false # no need to specify this since by default commands are asynchronous
	stdout:= !stdout # !stdout is the stdout for the whole script
	times:= 3 # run this 3 times
	every:= "1m" # run it every minute
}

# wait 30 seconds, we use ':' so it's synchronous
:sleep 30

@ :echo slept {
	stdout:= !stdout
}

# change the content of echoed.txt every minute by printing the hour
@ date +"%T" {
sync:= true # we don't need async here
	stdout:= echoed.txt
	every:= 1m
	times:= 2
}

#give tail command enough time to read the latest change
:sleep 31

# escape sequences are parsed if double quoted
@ :echo "\tI'm done" {
	stdout:= !stdout #shorthand for this coming soon
}
```

output:

```
$ ./example.ins
starting ./example.ins
hello
slept
"02:48:05"
"02:49:05"
        I'm done
done ./example.ins
```
