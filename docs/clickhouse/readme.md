# Clickhouse



Connecting to Clickhouse 

```bash

go run main.go --launch-vm=ch --preset=clickhouse --mem=8192 --cpu=4

sudo service clickhouse-server start


/etc/clickhouse-server/config.xml

<listen_host>0.0.0.0</listen_host>


sudo systemctl restart clickhouse-server

go run main.go --expose-vm=ch--port=8081 --hostport=8003 --external-ip=192.168.1.224 --protocol=tcp



```

```sql

CREATE TABLE my_first_table
(
    user_id UInt32,
    message String,
    timestamp DateTime,
    metric Float32
)
ENGINE = MergeTree
PRIMARY KEY (user_id, timestamp)



INSERT INTO my_first_table (user_id, message, timestamp, metric) VALUES
    (101, 'Hello, ClickHouse!',                                 now(),       -1.0    ),
    (102, 'Insert a lot of rows per batch',                     yesterday(), 1.41421 ),
    (102, 'Sort your data based on your commonly-used queries', today(),     2.718   ),
    (101, 'Granules are the smallest chunks of data read',      now() + 5,   3.14159 )



SELECT *
 FROM my_first_table
 ORDER BY timestamp


 
```