go-hpfeeds
==========

Basic Go client implementation of [hpfeeds](https://github.com/rep/hpfeeds), a simplistic
publish/subscribe protocol, written by Mark Schloesser ([rep](https://github.com/rep/)),
heavily used within the [Honeynet Project](https://honeynet.org/) for internal real-time
data sharing. Backend component of [Honeymap](https://github.com/fw42/honeymap) and
[hpfriends](http://hpfriends.honeycloud.net).

Usage
-----
See example and ```go doc```.

License
-------
BSD

TODO
----
* Test if everything actually works as intended, maybe write some unit tests 
* See if it's necessary to use more buffering for ```sendRawMsg()```
* Implement wrapper for JSON channels
* Write command line interface
* Add some sanity checks for message field length values
