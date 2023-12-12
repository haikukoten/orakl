# node:20.10.0-slim
FROM node@sha256:18aacc7993a16f1d766c21e3bff922e830bcdc7b549bbb789ceb7374a6138480 AS build

RUN apt-get update && apt-get install -y curl

WORKDIR /app

COPY package.json .

COPY yarn.lock .

COPY contracts contracts

COPY vrf vrf

COPY fetcher fetcher

COPY core core

RUN yarn core install

RUN yarn core build

FROM node@sha256:18aacc7993a16f1d766c21e3bff922e830bcdc7b549bbb789ceb7374a6138480

WORKDIR /app

RUN apt-get update && apt-get install -y curl

COPY --from=build /app/package.json /app/package.json

COPY --from=build /app/node_modules /app/node_modules

COPY --from=build /app/core/node_modules /app/core/node_modules

COPY --from=build /app/contracts /app/contracts

COPY --from=build /app/vrf /app/vrf

COPY --from=build /app/fetcher /app/fetcher

COPY --from=build /app/core /app/core

WORKDIR /app/core