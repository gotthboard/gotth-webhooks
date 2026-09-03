# Pinned address policy

The V1 candidate rejects every allocation in the IANA IPv4 and IPv6
Special-Purpose Address Registries, including entries whose `Global` field is
`True`. It does not reinterpret `Global` as ordinary application egress. IPv6
must additionally fall inside IANA's allocated global-unicast `2000::/3`.
Multicast and other non-global-unicast addresses are rejected independently.

Snapshot authority, fetched read-only on 2026-09-03:

- IPv4 registry, last updated 2025-10-09:
  `iana-ipv4-special-registry.xml`, SHA-256
  `cf24e11f41b7d42c68debe2d18b97cac815084ec413ebb3b244f704028a16f20`.
- IPv6 registry, last updated 2025-10-09:
  `iana-ipv6-special-registry.xml`, SHA-256
  `c17f4380ba84fb2160dae82ebfd8bd155a5853cfab624ed3a9fd251638a8be02`.
- Canonical URLs:
  `https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xml`
  and
  `https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xml`.

The production table uses these encompassing registry prefixes:

```text
IPv4: 0.0.0.0/8, 10.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8,
      169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24,
      192.31.196.0/24, 192.52.193.0/24, 192.88.99.0/24,
      192.168.0.0/16, 192.175.48.0/24, 198.18.0.0/15,
      198.51.100.0/24, 203.0.113.0/24, 240.0.0.0/4
IPv6: ::/128, ::1/128, ::ffff:0:0/96, 64:ff9b::/96,
      64:ff9b:1::/48, 100::/64, 100:0:0:1::/64, 2001::/23,
      2001:db8::/32, 2002::/16, 2620:4f:8000::/48, 3fff::/20,
      5f00::/16, fc00::/7, fe80::/10
```

Nested registry rows are covered by the registry's own broader `192.0.0.0/24`
and `2001::/23` rows. Tests duplicate the compact snapshot, require exact table
equality, and reject both endpoints of every prefix. There is no live registry
fetch at runtime. Maintainers must compare both registries and refresh this
snapshot before release and during every security maintenance cycle.
