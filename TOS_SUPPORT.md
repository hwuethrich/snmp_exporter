# TOS (Type of Service) / DSCP Support

The SNMP Exporter supports setting the IP Type of Service (TOS) / Differentiated Services Code Point (DSCP) value on SNMP packets. This feature allows you to prioritize SNMP traffic for Quality of Service (QoS) purposes in your network.

## Overview

The TOS/DSCP field in IP headers is used by network devices to classify and prioritize traffic. By setting a specific TOS value on SNMP packets, you can:

- Ensure SNMP monitoring traffic receives appropriate network priority
- Meet network QoS policies and requirements
- Improve reliability of SNMP monitoring in congested networks
- Comply with organizational network standards

## Usage

TOS can be set globally for all targets or per-target via URL parameter.

### Global Setting (Command-Line Flag)

Set a default TOS value for all SNMP targets:

```bash
./snmp_exporter --snmp.tos=184
```

### Per-Target Setting (URL Parameter)

Override the global TOS setting for specific targets using the `snmp_tos` URL parameter:

```
http://localhost:9116/snmp?target=192.0.0.8&snmp_tos=184
```

The URL parameter takes precedence over the command-line flag, allowing you to:
- Set different TOS values for different targets
- Override the global setting for specific high-priority devices
- Use the default (0) for some targets while setting TOS for others

### TOS Value Range

- Valid values: 0-255
- Default: 0 (no TOS is set)
- Value 0 disables TOS setting

### Common DSCP Values

Here are some commonly used DSCP values (multiply by 4 to get TOS value):

| DSCP Name | DSCP Value | TOS Value | Description |
|-----------|------------|-----------|-------------|
| CS0 (BE) | 0 | 0 | Best Effort (default) |
| CS1 | 8 | 32 | Low Priority |
| AF11 | 10 | 40 | High Throughput |
| AF21 | 18 | 72 | Low Latency |
| AF31 | 26 | 104 | Multimedia Streaming |
| AF41 | 34 | 136 | Multimedia Conferencing |
| CS5 | 40 | 160 | Signaling |
| EF | 46 | 184 | Expedited Forwarding (voice) |
| CS6 | 48 | 192 | Network Control |
| CS7 | 56 | 224 | Reserved |

**Note:** When using DSCP values, convert to TOS by multiplying by 4, or use the TOS value directly.

Example for Expedited Forwarding (EF):
```bash
./snmp_exporter --snmp.tos=184
```

## Platform Support

### Unix/Linux Systems
Full support for both IPv4 and IPv6:
- IPv4: Sets `IP_TOS` socket option
- IPv6: Sets `IPV6_TCLASS` socket option

### Other Platforms
On non-Unix platforms (e.g., Windows), TOS setting may not be supported. The exporter will:
- Accept `--snmp.tos=0` (no-op)
- Return an error for non-zero TOS values if the platform doesn't support it

## Configuration Examples

### Basic Usage (Global)
```bash
# Start with TOS value 184 (EF - Expedited Forwarding) for all targets
./snmp_exporter --snmp.tos=184 --config.file=snmp.yml
```

### Per-Target Usage with Prometheus
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'snmp-high-priority'
    static_configs:
      - targets:
        - 192.168.1.1  # Critical router
        - 192.168.1.2  # Critical switch
    metrics_path: /snmp
    params:
      module: [if_mib]
      snmp_tos: ['184']  # High priority for these targets
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9116

  - job_name: 'snmp-normal-priority'
    static_configs:
      - targets:
        - 192.168.2.1  # Regular device
    metrics_path: /snmp
    params:
      module: [if_mib]
      snmp_tos: ['0']  # Best effort for these targets
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9116
```

### With Other Options
```bash
# Combine global TOS with other common options
./snmp_exporter \
  --snmp.tos=184 \
  --config.file=/etc/snmp_exporter/snmp.yml \
  --web.listen-address=:9116 \
  --snmp.module-concurrency=2
```

### Docker
```bash
docker run -p 9116:9116 \
  -v /path/to/snmp.yml:/etc/snmp_exporter/snmp.yml \
  prom/snmp-exporter:latest \
  --snmp.tos=184
```

## Technical Details

The TOS feature works by:

1. Accepting a TOS value (0-255) via the `--snmp.tos` command-line flag
2. Validating the value is within the valid range
3. Setting the `Control` function on the underlying `gosnmp.GoSNMP` client
4. Using the `Control` function to set socket options via `setsockopt`:
   - `IP_TOS` for IPv4 sockets
   - `IPV6_TCLASS` for IPv6 sockets

The socket option is set before the SNMP connection is established, ensuring all SNMP packets use the specified TOS value.

## Verification

To verify that TOS is being set correctly, you can:

1. Check the startup log for the TOS value:
   ```
   level=INFO msg="Starting snmp_exporter" ... tos=184
   ```

2. Use network capture tools (tcpdump, Wireshark) to inspect the IP headers of SNMP packets:
   ```bash
   # Capture SNMP traffic and display TOS field
   tcpdump -i any -n -v 'port 161' | grep -i tos
   ```

3. In Wireshark, look at:
   - IPv4: IP header → Differentiated Services Field → DSCP
   - IPv6: IPv6 header → Traffic Class

## Troubleshooting

### Error: "TOS value must be between 0 and 255"
The provided TOS value is outside the valid range. Ensure you're using a value from 0 to 255.

### Error: "TOS socket option is not supported on this platform"
TOS setting is not supported on your operating system. This typically occurs on non-Unix platforms. You can still run the exporter with `--snmp.tos=0` (default).

### TOS not visible in network captures
- Ensure your network equipment and intermediate devices preserve the TOS/DSCP field
- Some routers or switches may rewrite or ignore TOS values based on their QoS policies
- Check that you have appropriate permissions to set socket options

## Best Practices

1. **Consult Network Team**: Before setting TOS values, coordinate with your network team to ensure the values align with your organization's QoS policies.

2. **Start Conservative**: Begin with Best Effort (0) and only increase priority if monitoring is impacted by network congestion.

3. **Monitor Impact**: After enabling TOS, monitor both the SNMP exporter's performance and overall network behavior.

4. **Document Configuration**: Keep track of which TOS values you're using and why.

5. **Test Thoroughly**: Test in a non-production environment first to ensure the TOS value doesn't cause unexpected network behavior.

## References

- [RFC 2474 - Definition of the Differentiated Services Field (DS Field)](https://tools.ietf.org/html/rfc2474)
- [RFC 2475 - An Architecture for Differentiated Services](https://tools.ietf.org/html/rfc2475)
- [RFC 3168 - The Addition of Explicit Congestion Notification (ECN) to IP](https://tools.ietf.org/html/rfc3168)
