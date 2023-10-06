// This example shows how to use the `meta_metrics_duration` transform to
// measure the execution time of other transforms.
local sub = import '../../../../../build/config/substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  transforms: [
    // The `meta_metrics_duration` transform measures the execution time of
    // the transform that it wraps.
    sub.transform.meta.metrics.duration(
      settings={
        name: 'ObjectCopyDuration',
        attributes: attr,
        destination: dest,
        transform: sub.transform.object.copy(
          settings={ object: { key: 'foo', set_key: 'baz' } },
        ),
      },
    ),
    // This can be useful for measuring the execution time of transforms that
    // may take a long time to execute. In this example, the `utility_delay`
    // transform is used to simulate a long-running transform.
    sub.transform.meta.metrics.duration(
      settings={
        name: 'UtilityDelayDuration',
        attributes: attr,
        destination: dest,
        transform: sub.transform.utility.delay(
          settings={ duration: '100ms' },
        ),
      },
    ),
    // Multiple transforms can be measured at once by wrapping them in a
    // `meta_pipeline` transform.
    sub.transform.meta.metrics.duration(
      settings={
        name: 'UtilityMultiDuration',
        attributes: attr,
        destination: dest,
        transform: sub.transform.meta.pipeline(
          settings={ transforms: [
            sub.transform.utility.delay(
              settings={ duration: '100ms' },
            ),
            sub.transform.utility.delay(
              settings={ duration: '100ms' },
            ),
            sub.transform.utility.delay(
              settings={ duration: '100ms' },
            ),
          ] }
        ),
      },
    ),

  ],
}
