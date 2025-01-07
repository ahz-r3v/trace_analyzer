# Trace_analyzer  
## usage  

**for azure2019**
```
go mod tidy  
go mod vendor  
go run main.go -wrapper=azure2019 -keepalive=60 -tolerance=100 -iatDistribution=1 -shiftIAT=false -granularity=0 <invocation_file_path> <duration_file_path> <output_file_path>  
```

**for azure2021**
```
go run main.go -wrapper=azure2021 -keepalive=60 -tolerance=100 <invocation_file_path> <output_file_path>  
```

## output format  
| time | ColdstartFrom0 | PeriodicInvocation |