# fiplog
[fiplog]() simple golang log lib with basic configurable log level, output format and location.

A typical config file could be:
```
level: Debug
file: fip.log
pattern: %date [%level] <%file> %msg
``` 

**Usage**
The usage is simple:
```
logger := fiplog.GetLogger()
logger.Debug("debug")
logger.Info("info")
logger.Error("error:",err)
```