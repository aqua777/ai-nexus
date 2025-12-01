FROM golang:1.25 AS golang

ARG GO_USER_ID=1001
ARG GO_USER_NAME=dev

ENV GO_USER_ID=${GO_USER_ID}
ENV GO_USER_NAME=${GO_USER_NAME}

ENV GOCACHE=/tmp/.cache/go/build
ENV GOMODCACHE=/tmp/.cache/go/pkg/mod

COPY --chmod=0755 .docker/scripts/go-* /usr/local/bin/

RUN apt-get update && apt-get install -y net-tools sqlite3 sudo zsh && \
    \
    groupadd -g "${GO_USER_ID}" "${GO_USER_NAME}" && \
    useradd -m -u "${GO_USER_ID}" -g "${GO_USER_ID}" "${GO_USER_NAME}" && \
    echo "${GO_USER_NAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    echo "User ${GO_USER_NAME} created with ID: ${GO_USER_ID}"

RUN go env
RUN bash /usr/local/bin/go-install-vscode-tools

# Download the two missing header files directly from the source
# this is a workaround for the missing headers in the arrow package
# for LanceDB to compile
RUN ARROW_URL="https://raw.githubusercontent.com/apache/arrow/apache-arrow-17.0.0/cpp/src/arrow/c" && \
    ARROW_TARGET_DIR="/usr/local/include/arrow/c" && \
    # Create the directory where the compiler expects the headers
    mkdir -p "${ARROW_TARGET_DIR}" && \
    # Download the two missing header files directly from the source
    curl -o "${ARROW_TARGET_DIR}/abi.h" "${ARROW_URL}/abi.h" && \
    curl -o "${ARROW_TARGET_DIR}/helpers.h" "${ARROW_URL}/helpers.h"

COPY .docker/zshrc /home/${GO_USER_NAME}/.zshrc

CMD ["/bin/bash"]
