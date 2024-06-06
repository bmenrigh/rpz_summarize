## *rpz_summarize* - eliminate redundant DNS entries for [RPZ](https://www.isc.org/rpz/)

When using wildcard policies it is easy to wind up with redundant entries in your zone. For example

    www.badguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.
    mail.badguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.
    ns1.badguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.
    *.badguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.

Here the `*.badguys.com` policy makes the other three redundant.

    $ ./rpz_summarize < tests/badguys.txt 
    *.badguys.com.rpz.example.com. IN 300 CNAME sinkhole.example.com.
    == ZONE STATS ==
    Lines: 4
    Unparsable: 0
    Policy targets: 1
    Entries in tree: 4
    Entries after summarization: 1 (75.00% reduction)

The zone is segmented by policy targets so for example sinkholed responses aren't compared with **NXDOMAIN** or **NODATA** or **RPZ-PASSTHRU**:

    *.shadyguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.
    wpad.shadyguys.com.rpz.example.com. 300 IN CNAME *.
    bad.shadyguys.com.rpz.example.com. 300 IN CNAME sinkhole.example.com.
    smtp.shadyguys.com.rpz.example.com. 300 IN CNAME rpz-passthru.
    *.static.shadyguys.com.rpz.example.com. 300 IN CNAME rpz-passthru.
    good.static.shadyguys.com.rpz.example.com. 300 IN CNAME rpz-passthru.

Which gets reduced like so:

    $ ./rpz_summarize < tests/passthru.txt 
    smtp.shadyguys.com.rpz.example.com. IN 300 CNAME rpz-passthru.
    *.static.shadyguys.com.rpz.example.com. IN 300 CNAME rpz-passthru.
    *.shadyguys.com.rpz.example.com. IN 300 CNAME sinkhole.example.com.
    wpad.shadyguys.com.rpz.example.com. IN 300 CNAME *.
    == ZONE STATS ==
    Lines: 6
    Unparsable: 0
    Policy targets: 3
    Entries in tree: 6
    Entries after summarization: 4 (33.33% reduction)


## Speed

With a moderately large zone (720k entries) the code takes around 1 second:

    $ time ./rpz_summarize < /tmp/rpz.txt > /dev/null
    == ZONE STATS ==
    Lines: 722724
    Unparsable: 0
    Policy targets: 6
    Entries in tree: 722724
    Entries after summarization: 554819 (23.23% reduction)
    
    real	0m1.071s
    user	0m2.484s
    sys	0m0.125s

Peak memory usage while processing this zone is about 160MB.
