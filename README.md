## ddns

Simple server for accepting DDNS requests and updating the matching records in
Route53 with authentication.

### Configuration

#### IAM Permission Policy

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": "*"
        }
    ]
}
```

### Usage

#### pfsense

Under `Services > Dynamic DNS > Dynamic DNS Clients`, click `Add`.

| Setting | Value |
| :---: | --- |
| Service Type | `Custom` |
| Username | `username` |
| Password | `password` |
| Update URL | `https://example.com?host=dynamic.example.com&ip=%IP%` |
| Result Match | `{"status":"ok","ip":"%IP%"}` |
