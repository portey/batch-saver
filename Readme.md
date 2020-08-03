# batch-saver

### How to run
Just enter `make docker-up`

### Choosing database
I have implemented store as an abstraction and along with it, I have to implementations of it - Redis and Postgres. 
Depending on data usage I would use the different database: 
- Redis - if we have huge rate, a lot of RAM,  need expiration or if data is not super important and may leave not forever.
- PostgreSQL - if we should save data forever