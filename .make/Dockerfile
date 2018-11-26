FROM centos:7
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Devtools <devtools@redhat.com>"
ENV LANG=en_US.utf8

# Create a non-root user and a group with the same name: "build"
ENV F8_USER_NAME=build
RUN useradd --no-create-home -s /bin/bash ${F8_USER_NAME}

# Install little pcp pmcd server for metrics collection
# would prefer only pmcd, and not the /bin/pm*tools etc.
COPY .make/pcp/pcp.repo /etc/yum.repos.d/pcp.repo
RUN yum install -y pcp pcp-pmda-prometheus && yum clean all && \
    mkdir -p /etc/pcp /var/run/pcp /var/lib/pcp /var/log/pcp  && \
    chgrp -R root /etc/pcp /var/run/pcp /var/lib/pcp /var/log/pcp && \
    chmod -R g+rwX /etc/pcp /var/run/pcp /var/lib/pcp /var/log/pcp
COPY .make/pcp/fabric8-build+pmcd.sh /fabric8-build+pmcd.sh
EXPOSE 44321

COPY bin/fabric8-build /bin/fabric8-build

# From here onwards, any RUN, CMD, or ENTRYPOINT will be run under the following user
USER ${F8_USER_NAME}

WORKDIR /
ENTRYPOINT [ "/fabric8-build+pmcd.sh" ]

EXPOSE 8080
