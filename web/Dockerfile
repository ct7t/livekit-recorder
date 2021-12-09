FROM ubuntu:20.10
RUN apt update
RUN apt install -y curl

# Install npm / Node
ENV NODE_VERSION=14.17.5
ENV NVM_DIR=/root/.nvm
RUN mkdir /root/.nvm
RUN curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.38.0/install.sh | bash
RUN . "$NVM_DIR/nvm.sh" && nvm install ${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm use v${NODE_VERSION}
RUN . "$NVM_DIR/nvm.sh" && nvm alias default v${NODE_VERSION}
ENV PATH="/root/.nvm/versions/node/v${NODE_VERSION}/bin/:${PATH}"
RUN node --version
RUN npm --version

RUN mkdir livekit-template
COPY ./ livekit-template/
WORKDIR livekit-template/build

CMD [ "npx", "http-server" ]
