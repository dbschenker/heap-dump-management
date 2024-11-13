# Heap Dump Management

As the filesystem in Kubernetes pods is not directly accessable we developed a sidecar approach where written heap dumps are collected and stored in an encrypted format on a central S3 Bucket per cluster.

This approach consists of 3 parts: 

* [heap dump service](heap-dump-service/README.md)
* [notify sidecar](notify-sidecar/README.md)
* [heap dump companion](heap-dump-companion/README.md)

Please read the individual documentation to understand the individual parts.  
This architectual overview should help to understand the interactions in between the components

![](notify-sidecar/docs/Architecture.svg)

## Maintainers

This project is maintained by: 
* Tobias Trabelsi (tobias.trabelsi+github@dbschenker.com)
