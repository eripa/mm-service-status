# mm-service-status

MediaMarkt Service Status Check tool

**Note:** Currently only designed for Swedish store

Can be used in conjunction with crontab + grep and something like Pushover to notify on service status change

# Usage

    go get github.com/eripa/mm-service-status

    mm-service-status  --lastname "Ripa" --order-id 128732 --store-id 1157

    Checking service status for order: 128732 name: Ripa store: 1157
    Name: ERIC RIPA
    Product: MY AWESOME PRODUCT
    Status: Ärendet hanteras

# License

MIT, see LICENSE file for full details
