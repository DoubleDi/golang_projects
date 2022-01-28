# BTC Billionaire

## Required
1. docker
2. docker-compose
3. make

## Run 
1. clone the repo
```
git clone git@github.com:DoubleDi/golang_projects.git
```

2. go to billionaire
```
cd golang_projects/billionaire
```

3. start the app
```
make run
```
the app will start on http://127.0.0.1:8000
the database port is :5432

4. configuration 
all the configs are in configs/

## API
```
POST /balance
GET /balance
```