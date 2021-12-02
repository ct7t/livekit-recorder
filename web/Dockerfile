FROM ubuntu:20.10
RUN apt update && apt install -y curl

# Install npm / Node
ENV NODE_VERSION=14.17.5
ENV NVM_DIR=/root/.nvm
RUN mkdir ${NVM_DIR} && curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.38.0/install.sh | bash
RUN . "$NVM_DIR/nvm.sh" && nvm install ${NODE_VERSION} && nvm use v${NODE_VERSION} && nvm alias default v${NODE_VERSION}
ENV PATH="/root/.nvm/versions/node/v${NODE_VERSION}/bin/:${PATH}"

WORKDIR livekit-template
COPY ./build/ .

CMD [ "npx", "http-server" ]
