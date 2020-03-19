FROM alpine as runtime
ADD ["coredns", "/bin/coredns"]
EXPOSE 53 53/udp
ENTRYPOINT ["/bin/coredns"]