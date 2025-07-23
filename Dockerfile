
# this is the build stage
FROM golang:1.23.11-alpine3.22 AS builder

# Set working directory inside the container to /app
WORKDIR /app

# Copy all files from current host directory into /app in container
COPY . .
# Compile the Go app to a single binary called "main"
# Syntax: go build -o <output_binary> <entry_point>
RUN go build -o main main.go

# Install curl because Alpine doesn't include it by default
# curl is needed to download the migrate binary
RUN apk add --no-cache curl

# Download and extract the golang-migrate CLI binary
# This tool is used to apply DB schema migrations
RUN curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz

# run stage 
# same apline verison to ensure compatibility
# note alpine : VERSION
FROM alpine:3.22 
WORKDIR /app

# copy the executable bin exe from buildstage(from=builder[see it's name]) to runstage
# i.e inside the builder we are copying from app/main to "."is CWD
# Copy the compiled Go binary from the builder stage into the current image
# Source: /app/main in the builder stage
# Destination: current working directory (/app) in this stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate

# we missed the config file so it fails
# we also have to copy all dbmigration files into the stage
COPY app.env .
COPY start.sh /app/start.sh
COPY wait-for.sh .
RUN chmod +x /app/start.sh

COPY db/migration ./migration

#* added to make the sh file an executable one


# conatiner listens to thisport
EXPOSE 8081

# last step is default command when container starts
CMD [ "/app/main" ]
# when CMD is used with entry point it willact as additional @params
# can also be written as ENTRYPOINT [ "/app/start.sh" , "/app/main"]
# but this method gives flexibity
#! since we have to run the migrations before the start of the app we use the start.sh and beore
# ! we have to make the sh file an exe
ENTRYPOINT [ "/app/start.sh" ]


# FROM alpine	Small, efficient base image for runtime
# WORKDIR /app	All relative paths will start here
# COPY --from=builder /app/main .	Copy the built Go executable
# COPY --from=builder /app/migrate ./migrate	Copy the migrate CLI tool if needed inside container
# COPY app.env .	Environment variables needed at runtime
# COPY db/migration ./migration	SQL files needed for DB setup or versioning