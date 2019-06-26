# premconverter

## What is it?

If you've been using Premiere for a while, especially in a corporate environment, you'll know about the pain
of upgrading.  At The Guardian, we closely manage all our project files and this means that simply having
the users "save as" a project is not an option, because the project file can then easily get lost and they
then lose their work.

Fortunately, (at least when upgrading from 2017 to 2019) the underlying data structures within the file
are the same and we can simply update the version identifier within the file to allow the later Premiere
to open a project file that was saved in the earlier version.

This app automates that process, allowing you to efficiently batch-convert many project files from the older
to the newer version.

## What does it run on?

The app runs from the commandline on any platform that Go supports - including MS Windows, Mac and Linux in 32 and 64 bit flavours.

## DISCLAIMER

This process involves messing with the internal content of the project file, which Adobe specifically do not 
recommend.  Neither I as the author nor Adobe can be held responsible if this operation corrupts your
project file. **ALWAYS**, **always**, **ALWAYS** have a backup before messing with your project file in this
way; if anything goes wrong then you can always get back to where you started.
I would recommend testing the converted file throughly in Premiere, just because it opens does not mean
that there is not some kind of subtle incompatibility hiding in there that may crash your Premiere.

I repeat: **make sure your project file is backed up before messing around with it**.

## How do I build it?

Step 1 - ensure you have Go v1.11 or later installed. Also ensure you have GNU make installed (Mac and Linux)
should have this already

Step 2 - clone this repo and run `make test && make` from the root of the checkout

Step 3 - this should run all the tests and output the compiled program to the `bin/` directory. Choose the appropriate
one for your operating system

Step 4 - copy the relevant binary to somewhere on your PATH, e.g. `cp bin/premconverter.macos /usr/local/bin` for Mac.

### But I'm on Windows!

Ah.  Well, unfortunately, I don't do any development on Windows so I'm not best-placed to help.
My recommendation would be to install Docker, then run (from a command prompt):

```console
docker run --rm -it -v {path-to-checkout}:/usr/src golang:1.12-stretch
[wait for Docker to download the image and put you to a prompt....]
cd /usr/src
make test && make
exit
```

This allows you to fairly simply have a working Linux environment to perform the build.
Alternatively, of course, you can just follow the instructions to set up Go on Windows - I have not
done this myself, so I can't presume to offer any advice on it though.

## Do I HAVE to build it myself?

Right now, I'm afraid yes - I am in the process of setting up an automated build and when that is done will update
this doc with details of where you can download precompiled binaries from.

## How do I run it?

With that part out of the way, now comes the simple part:

```console
$ premconverter.macos --help
Usage of bin/premconverter.macos:
  -input string
    	a single prproj file to process
  -output string
    	a single prproj file to output
    	
$ premconverter.macos -input /path/to/my/current.prproj -output /path/to/my/updated.prproj
```
The last line will update `current.prproj` into a new file at `updated.prproj`, leaving `current.prproj`
unmodified.  It can take a while if the project file is large.

You can then attempt to open `updated.prproj` in Premiere 2019.

## Wait, you said something about batching?

Yes, that is in development at the moment.

