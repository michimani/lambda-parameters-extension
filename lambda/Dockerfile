FROM public.ecr.aws/lambda/provided:al2 as build
# install build tools
RUN yum install -y golang unzip
RUN go env -w GOPROXY=direct
# cache dependencies
ADD go.mod go.sum ./
RUN go mod download
# build
ADD . .
RUN go build -o /main
# install extension
RUN mkdir -p /opt
ADD ./extension/ps-ex.zip ./
RUN unzip ps-ex.zip -d /opt
RUN rm ps-ex.zip
# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /main /main
# COPY entry.sh /
# RUN chmod 755 /entry.sh
RUN mkdir -p /opt/extensions
WORKDIR /opt/extensions
COPY --from=build /opt/extensions .
ENV ENV prod
# ENTRYPOINT [ "/entry.sh" ]
ENTRYPOINT [ "/main" ]
# CMD ["/main"]