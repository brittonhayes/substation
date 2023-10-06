// This example shows how to use the `utility_metrics_count` transform to
// count the number of messages received and transformed by Substation.
local sub = import '../../../../../build/config/substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  transforms: [
    // If the transform is configured first, then the count reflects
    // the number of messages received by Substation.
    sub.transform.utility.metrics.count(
      settings={ name: 'MessagesReceived', attributes: attr, destination: dest },
    ),
    sub.transform.utility.drop(),
    // If the transform is configured last, then the count reflects
    // the number of messages transformed by Substation.
    sub.transform.utility.metrics.count(
      settings={ name: 'MessagesTransformed', attributes: attr, destination: dest },
    ),
  ],
}
