FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bootstrap ./cmd

FROM public.ecr.aws/lambda/provided:al2023 AS runtime
WORKDIR ${LAMBDA_TASK_ROOT}
# Custom runtime bootstrap must exist under /var/runtime.
COPY --from=build /out/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap

# 默认执行 root command（即 Lambda runtime），也可覆盖成 dev/test-event
CMD ["bootstrap"]
