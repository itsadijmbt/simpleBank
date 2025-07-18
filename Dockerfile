
# this is the build stage
FROM golang:1.23.11-alpine3.22 AS builder 
# declare the working directory inside the image
WORKDIR /app
# from and destination "." everything form current wher we run dok boot for image
# and "." is the CWD ie current workingdir where the files/images are being copied to
COPY . .
# building our APP to a single binary exe
# go build -o __name of exe__  __entry point___ 
RUN go build -o main main.go



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
# we missed the config file so it fails
# TODO : production config 
COPY app.env .


# conatiner listens to thisport
EXPOSE 8081
# last step is default command when container starts
CMD [ "/app/main" ]
 