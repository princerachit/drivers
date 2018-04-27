FROM alpine
LABEL maintainers="Kubernetes Authors"
LABEL description="HostPath Driver"

#ADD ./_output/hostpathplugin /hostpathplugin
COPY ./_output/hostpathplugin /hostpathplugin
ENTRYPOINT ["/hostpathplugin"]
