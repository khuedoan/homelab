# Prepare

Role intended to setup servers: login, packages, disks...

## Optional variables

```yaml
# valid values from `/sys/devices/system/cpu/cpu{0,1,2,3}/cpufreq/scaling_available_frequencies`
# e.g. for rock64: 408000 600000 816000 1008000 1200000 1296000 1392000 1512000
prepare_max_cpu_freq=1512000
```
