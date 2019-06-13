# Humanize

Format your [logrus](https://github.com/sirupsen/logrus) logs for Human Eyes.

Spits out logs like-a this:
 
``` shell
// --- Long Format --- //

2018-11-08T10:49:28 [info]: This is the very polite log message
          Fields:
                    dance:         flhargunstow
                    power_level:   9000
2018-11-08T10:49:28 [error]: Alas, error city!
          Fields:
                    dance:         flhargunstow
                    power_level:   9000
Error: oh heavens oh no an error eep

// --- Compact Format --- //

2018-11-08T10:49:28 [info]: Now, compact!
        	power_level: 9000	dance: flhargunstow
2018-11-08T10:49:28 [error]: This is a very compact error message
        	dance: flhargunstow	power_level: 9000
Error: oh heavens oh no an error eep
```


