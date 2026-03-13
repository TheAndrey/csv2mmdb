# CSV to mmdb

Generates MaxMind MMDB database from CSV file.

## Example

Input CSV files (routes + as names)

```
asn,as_name
10000,"Nagasaki Cable Media Inc."
100,"FMC Central Engineering Laboratories"
10,"CSNET Coordination and Information Center (CSNET-CIC)"
1,"Level 3 Parent, LLC"
```

```
cidr,asn
14.1.8.0/21,10000
38.22.219.0/24,100
195.74.62.0/23,10
203.109.36.0/24,1
```

Database record:

```json
{
  "route": "195.74.62.0/23",
  "asn": 10,
  "as_name": "CSNET Coordination and Information Center (CSNET-CIC)"
}
```