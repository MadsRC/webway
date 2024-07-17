# Webway

Webway is an Apache Kafka® compatible data streaming platform, built from the
ground to be a distributed, scalable, fault-tolerant, and highly available
platform for processing and storing data streams. In contrast to Apache Kafka®,
there is no local storage to worry about and no need for cross-AZ replication.

The good folks over at [WarpStream] have done a great job of building a solid
platform for streaming data. Webway takes heavy inspiration from their project,
and references many of their public design documents. Despite taking heavy
inspiration from [WarpStream], Webway is not a fork of [WarpStream]. Webway is
a completely new project, and is not affiliated with [WarpStream] in any way.

## Related external design documents

- https://cwiki.apache.org/confluence/display/KAFKA/A+Guide+To+The+Kafka+Protocol#AGuideToTheKafkaProtocol-CommonRequestandResponseStructure
- https://docs.warpstream.com/warpstream/overview/architecture
- https://docs.warpstream.com/warpstream/overview/architecture/service-discovery
- https://www.warpstream.com/blog/hacking-the-kafka-protocol
- https://docs.warpstream.com/warpstream/overview/architecture/write-path
- https://www.warpstream.com/blog/unlocking-idempotency-with-retroactive-tombstones
- https://docs.warpstream.com/warpstream/overview/architecture/read-path
- https://www.warpstream.com/blog/minimizing-s3-api-costs-with-distributed-mmap
- https://docs.warpstream.com/warpstream/overview/architecture/life-of-a-request

[WarpStream]: https://warpstream.com