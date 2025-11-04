# Dockerfile for Trust Vault Plugin with HashiCorp Vault
# This uses Dockerfile.build to get Trust Wallet Core libraries and the plugin binary
# The plugin binary and TWC libraries are built in Dockerfile.build

# Stage 1: Build TWC and plugin (from Dockerfile.build)
FROM ubuntu:22.04 AS wallet-core-builder
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    build-essential cmake git libboost-all-dev ninja-build \
    clang-14 llvm-14 libc++-dev libc++abi-dev pkg-config \
    libtool autoconf ruby-full wget curl ca-certificates dos2unix \
    && rm -rf /var/lib/apt/lists/*
ENV CC=clang-14
ENV CXX=clang++-14
RUN wget "https://sh.rustup.rs" -O rustup.sh && sh rustup.sh -y && rm rustup.sh
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup default nightly && rustup update nightly
RUN cargo install cbindgen --locked
WORKDIR /build
COPY third_party/wallet-core /build/wallet-core
WORKDIR /build/wallet-core
RUN set -e && \
    find . -type f \( -name "*.sh" -o -name "*.rb" -o -path "*/tools/*" -o -path "*/codegen/bin/*" \) ! -path "*/build/*" ! -path "*/.git/*" -exec sh -c 'dos2unix "{}" 2>/dev/null || sed -i "s/\r$//" "{}"' \; && \
    chmod +x tools/* codegen/bin/* 2>/dev/null || true && \
    bash tools/install-dependencies && \
    PATH="/root/.cargo/bin:${PATH}" bash tools/generate-files native && \
    cmake -H. -Bbuild -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_COMPILER=clang++-14 -DCMAKE_C_COMPILER=clang-14 -GNinja && \
    cmake --build build --parallel 4 --target TrustWalletCore

FROM ubuntu:22.04 AS plugin-builder
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    wget build-essential clang-14 llvm-14 libc++-dev libc++abi-dev \
    libboost-all-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then GO_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then GO_ARCH="arm64"; \
    else echo "Unsupported: $ARCH" && exit 1; fi && \
    wget https://go.dev/dl/go1.23.4.linux-${GO_ARCH}.tar.gz && \
    tar -C /usr/local -xzf go1.23.4.linux-${GO_ARCH}.tar.gz && rm go1.23.4.linux-${GO_ARCH}.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}" GOROOT="/usr/local/go"
COPY --from=wallet-core-builder /build/wallet-core/build/libTrustWalletCore.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/build/trezor-crypto/libTrezorCrypto.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/rust/target/release/libwallet_core_rs.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/build/local/lib/libprotobuf*.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/include /usr/local/include
COPY --from=wallet-core-builder /build/wallet-core/build/local/include /usr/local/include
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=1 CC=clang-14 CXX=clang++-14
ENV CGO_CFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib -lTrustWalletCore -lwallet_core_rs -lTrezorCrypto -lprotobuf -lstdc++ -lm -lpthread"
RUN go build -o trust-vault-plugin cmd/trust-vault/main.go

# Stage 2: Vault image with TWC libraries and plugin
FROM ubuntu:22.04

# Install Vault (download binary directly)
RUN apt-get update && apt-get install -y \
    wget curl ca-certificates unzip \
    && rm -rf /var/lib/apt/lists/*
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then VAULT_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then VAULT_ARCH="arm64"; \
    else echo "Unsupported: $ARCH" && exit 1; fi && \
    wget https://releases.hashicorp.com/vault/1.15.6/vault_1.15.6_linux_${VAULT_ARCH}.zip && \
    unzip vault_1.15.6_linux_${VAULT_ARCH}.zip && \
    mv vault /usr/local/bin/vault && \
    chmod +x /usr/local/bin/vault && \
    rm vault_1.15.6_linux_${VAULT_ARCH}.zip && \
    vault version

# Install runtime dependencies for Trust Wallet Core (C++ standard library)
RUN apt-get update && apt-get install -y \
    libstdc++6 libgcc-s1 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy Trust Wallet Core libraries from builder
COPY --from=wallet-core-builder /build/wallet-core/build/libTrustWalletCore.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/build/trezor-crypto/libTrezorCrypto.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/rust/target/release/libwallet_core_rs.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/build/local/lib/libprotobuf*.a /usr/local/lib/
COPY --from=wallet-core-builder /build/wallet-core/include /usr/local/include
COPY --from=wallet-core-builder /build/wallet-core/build/local/include /usr/local/include

# Copy plugin binary from plugin builder
COPY --from=plugin-builder /workspace/trust-vault-plugin /vault/plugins/trust-vault-plugin

# Create plugin directory and set permissions
RUN mkdir -p /vault/plugins && \
    chmod +x /vault/plugins/trust-vault-plugin && \
    sha256sum /vault/plugins/trust-vault-plugin | awk '{print $1}' > /vault/plugins/trust-vault-plugin.sha256

# Create vault user (Vault runs as this user)
RUN useradd -r -d /vault -s /bin/false vault && \
    chown -R vault:vault /vault && \
    chmod 755 /vault/plugins && \
    chmod 755 /vault/plugins/trust-vault-plugin
