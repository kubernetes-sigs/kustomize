#!/usr/bin/env bash
#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------
#
# Docs: https://github.com/microsoft/vscode-dev-containers/blob/main/script-library/docs/docker.md
# Maintainer: The VS Code and Codespaces Teams

DOCKER_VERSION="${VERSION:-"latest"}"
USE_MOBY="${MOBY:-"true"}"
MOBY_BUILDX_VERSION="${MOBYBUILDXVERSION:-"latest"}"
DOCKER_DASH_COMPOSE_VERSION="${DOCKERDASHCOMPOSEVERSION:-"v2"}" # v1 or v2 or none

ENABLE_NONROOT_DOCKER="${ENABLE_NONROOT_DOCKER:-"true"}"
SOURCE_SOCKET="${SOURCE_SOCKET:-"/var/run/docker-host.sock"}"
TARGET_SOCKET="${TARGET_SOCKET:-"/var/run/docker.sock"}"
USERNAME="${USERNAME:-"${_REMOTE_USER:-"automatic"}"}"
INSTALL_DOCKER_BUILDX="${INSTALLDOCKERBUILDX:-"true"}"
INSTALL_DOCKER_COMPOSE_SWITCH="${INSTALLDOCKERCOMPOSESWITCH:-"true"}"
MICROSOFT_GPG_KEYS_URI="https://packages.microsoft.com/keys/microsoft.asc"
MICROSOFT_GPG_KEYS_ROLLING_URI="https://packages.microsoft.com/keys/microsoft-rolling.asc"
DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES="trixie bookworm buster bullseye bionic focal jammy noble plucky"
DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES="trixie bookworm buster bullseye bionic focal hirsute impish jammy noble plucky"

set -e

# Clean up
rm -rf /var/lib/apt/lists/*

# Setup STDERR.
err() {
    echo "(!) $*" >&2
}

if [ "$(id -u)" -ne 0 ]; then
    echo -e 'Script must be run as root. Use sudo, su, or add "USER root" to your Dockerfile before running this script.'
    exit 1
fi

# Determine the appropriate non-root user
if [ "${USERNAME}" = "auto" ] || [ "${USERNAME}" = "automatic" ]; then
    USERNAME=""
    POSSIBLE_USERS=("vscode" "node" "codespace" "$(awk -v val=1000 -F ":" '$3==val{print $1}' /etc/passwd)")
    for CURRENT_USER in "${POSSIBLE_USERS[@]}"; do
        if id -u ${CURRENT_USER} > /dev/null 2>&1; then
            USERNAME=${CURRENT_USER}
            break
        fi
    done
    if [ "${USERNAME}" = "" ]; then
        USERNAME=root
    fi
elif [ "${USERNAME}" = "none" ] || ! id -u ${USERNAME} > /dev/null 2>&1; then
    USERNAME=root
fi

apt_get_update()
{
    if [ "$(find /var/lib/apt/lists/* | wc -l)" = "0" ]; then
        echo "Running apt-get update..."
        apt-get update -y
    fi
}

# Checks if packages are installed and installs them if not
check_packages() {
    if ! dpkg -s "$@" > /dev/null 2>&1; then
        apt_get_update
        apt-get -y install --no-install-recommends "$@"
    fi
}

# Figure out correct version of a three part version number is not passed
find_version_from_git_tags() {
    local variable_name=$1
    local requested_version=${!variable_name}
    if [ "${requested_version}" = "none" ]; then return; fi
    local repository=$2
    local prefix=${3:-"tags/v"}
    local separator=${4:-"."}
    local last_part_optional=${5:-"false"}
    if [ "$(echo "${requested_version}" | grep -o "." | wc -l)" != "2" ]; then
        local escaped_separator=${separator//./\\.}
        local last_part
        if [ "${last_part_optional}" = "true" ]; then
            last_part="(${escaped_separator}[0-9]+)?"
        else
            last_part="${escaped_separator}[0-9]+"
        fi
        local regex="${prefix}\\K[0-9]+${escaped_separator}[0-9]+${last_part}$"
        local version_list="$(git ls-remote --tags ${repository} | grep -oP "${regex}" | tr -d ' ' | tr "${separator}" "." | sort -rV)"
        if [ "${requested_version}" = "latest" ] || [ "${requested_version}" = "current" ] || [ "${requested_version}" = "lts" ]; then
            declare -g ${variable_name}="$(echo "${version_list}" | head -n 1)"
        else
            set +e
            declare -g ${variable_name}="$(echo "${version_list}" | grep -E -m 1 "^${requested_version//./\\.}([\\.\\s]|$)")"
            set -e
        fi
    fi
    if [ -z "${!variable_name}" ] || ! echo "${version_list}" | grep "^${!variable_name//./\\.}$" > /dev/null 2>&1; then
        echo -e "Invalid ${variable_name} value: ${requested_version}\nValid values:\n${version_list}" >&2
        exit 1
    fi
    echo "${variable_name}=${!variable_name}"
}

# Use semver logic to decrement a version number then look for the closest match
find_prev_version_from_git_tags() {
    local variable_name=$1
    local current_version=${!variable_name}
    local repository=$2
    # Normally a "v" is used before the version number, but support alternate cases
    local prefix=${3:-"tags/v"}
    # Some repositories use "_" instead of "." for version number part separation, support that
    local separator=${4:-"."}
    # Some tools release versions that omit the last digit (e.g. go)
    local last_part_optional=${5:-"false"}
    # Some repositories may have tags that include a suffix (e.g. actions/node-versions)
    local version_suffix_regex=$6
    # Try one break fix version number less if we get a failure. Use "set +e" since "set -e" can cause failures in valid scenarios.
    set +e
        major="$(echo "${current_version}" | grep -oE '^[0-9]+' || echo '')"
        minor="$(echo "${current_version}" | grep -oP '^[0-9]+\.\K[0-9]+' || echo '')"
        breakfix="$(echo "${current_version}" | grep -oP '^[0-9]+\.[0-9]+\.\K[0-9]+' 2>/dev/null || echo '')"

        if [ "${minor}" = "0" ] && [ "${breakfix}" = "0" ]; then
            ((major=major-1))
            declare -g ${variable_name}="${major}"
            # Look for latest version from previous major release
            find_version_from_git_tags "${variable_name}" "${repository}" "${prefix}" "${separator}" "${last_part_optional}"
        # Handle situations like Go's odd version pattern where "0" releases omit the last part
        elif [ "${breakfix}" = "" ] || [ "${breakfix}" = "0" ]; then
            ((minor=minor-1))
            declare -g ${variable_name}="${major}.${minor}"
            # Look for latest version from previous minor release
            find_version_from_git_tags "${variable_name}" "${repository}" "${prefix}" "${separator}" "${last_part_optional}"
        else
            ((breakfix=breakfix-1))
            if [ "${breakfix}" = "0" ] && [ "${last_part_optional}" = "true" ]; then
                declare -g ${variable_name}="${major}.${minor}"
            else
                declare -g ${variable_name}="${major}.${minor}.${breakfix}"
            fi
        fi
    set -e
}

# Function to fetch the version released prior to the latest version
get_previous_version() {
    local url=$1
    local repo_url=$2
    local variable_name=$3
    prev_version=${!variable_name}

    output=$(curl -s "$repo_url");

    check_packages jq

    if echo "$output" | jq -e 'type == "object"' > /dev/null; then
        message=$(echo "$output" | jq -r '.message')
        if [[ $message == "API rate limit exceeded"* ]]; then
            echo -e "\nAn attempt to find previous to latest version using GitHub Api Failed... \nReason: ${message}"
            echo -e "\nAttempting to find previous to latest version using GitHub tags."
            find_prev_version_from_git_tags prev_version "$url" "tags/v"
            declare -g ${variable_name}="${prev_version}"
        fi
    elif echo "$output" | jq -e 'type == "array"' > /dev/null; then
        echo -e "\nAttempting to find previous version using GitHub Api."
        version=$(echo "$output" | jq -r '.[1].tag_name')
        declare -g ${variable_name}="${version#v}"
    fi
    echo "${variable_name}=${!variable_name}"
}

get_github_api_repo_url() {
    local url=$1
    echo "${url/https:\/\/github.com/https:\/\/api.github.com\/repos}/releases"
}

install_compose_switch_fallback() {
    compose_switch_url=$1
    repo_url=$(get_github_api_repo_url "${compose_switch_url}")
    echo -e "\n(!) Failed to fetch the latest artifacts for compose-switch v${compose_switch_version}..."
    get_previous_version "${compose_switch_url}" "${repo_url}" compose_switch_version
    echo -e "\nAttempting to install v${compose_switch_version}"
    curl -fsSL "https://github.com/docker/compose-switch/releases/download/v${compose_switch_version}/docker-compose-linux-${architecture}" -o /usr/local/bin/compose-switch
}

# Ensure apt is in non-interactive to avoid prompts
export DEBIAN_FRONTEND=noninteractive

# Install dependencies
check_packages apt-transport-https curl ca-certificates gnupg2 dirmngr wget
if ! type git > /dev/null 2>&1; then
    check_packages git
fi

# Source /etc/os-release to get OS info
. /etc/os-release
# Fetch host/container arch.
architecture="$(dpkg --print-architecture)"

# Prevent attempting to install Moby on Debian trixie (packages removed)
if [ "${USE_MOBY}" = "true" ] && [ "${ID}" = "debian" ] && [ "${VERSION_CODENAME}" = "trixie" ]; then
    err "The 'moby' option is not supported on Debian 'trixie' because 'moby-cli' and related system packages have been removed from that distribution."
    err "To continue, either set the feature option '\"moby\": false' or use a different base image (for example: 'debian:bookworm' or 'ubuntu-24.04')."
    exit 1
fi

# Check if distro is supported
if [ "${USE_MOBY}" = "true" ]; then
    if [[ "${DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES}" != *"${VERSION_CODENAME}"* ]]; then
        err "Unsupported  distribution version '${VERSION_CODENAME}'. To resolve, either: (1) set feature option '\"moby\": false' , or (2) choose a compatible OS distribution"
        err "Supported distributions include:  ${DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES}"
        exit 1
    fi
    echo "Distro codename  '${VERSION_CODENAME}'  matched filter  '${DOCKER_MOBY_ARCHIVE_VERSION_CODENAMES}'"
else
    if [[ "${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}" != *"${VERSION_CODENAME}"* ]]; then
        err "Unsupported distribution version '${VERSION_CODENAME}'. To resolve, please choose a compatible OS distribution"
        err "Supported distributions include:  ${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}"
        exit 1
    fi
    echo "Distro codename  '${VERSION_CODENAME}'  matched filter  '${DOCKER_LICENSED_ARCHIVE_VERSION_CODENAMES}'"
fi

# Set up the necessary apt repos (either Microsoft's or Docker's)
if [ "${USE_MOBY}" = "true" ]; then

    cli_package_name="moby-cli"

    # Import key safely and import Microsoft apt repo
    {
        curl -sSL ${MICROSOFT_GPG_KEYS_URI}
        curl -sSL ${MICROSOFT_GPG_KEYS_ROLLING_URI}
    } | gpg --dearmor > /usr/share/keyrings/microsoft-archive-keyring.gpg
    echo "deb [arch=${architecture} signed-by=/usr/share/keyrings/microsoft-archive-keyring.gpg] https://packages.microsoft.com/repos/microsoft-${ID}-${VERSION_CODENAME}-prod ${VERSION_CODENAME} main" > /etc/apt/sources.list.d/microsoft.list
else
    # Name of proprietary engine package
    cli_package_name="docker-ce-cli"

    # Import key safely and import Docker apt repo
    curl -fsSL https://download.docker.com/linux/${ID}/gpg | gpg --dearmor > /usr/share/keyrings/docker-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/${ID} ${VERSION_CODENAME} stable" > /etc/apt/sources.list.d/docker.list
fi

# Refresh apt lists
apt-get update

# Soft version matching for CLI
if [ "${DOCKER_VERSION}" = "latest" ] || [ "${DOCKER_VERSION}" = "lts" ] || [ "${DOCKER_VERSION}" = "stable" ]; then
    # Empty, meaning grab whatever "latest" is in apt repo
    cli_version_suffix=""
else
    # Fetch a valid version from the apt-cache (eg: the Microsoft repo appends +azure, breakfix, etc...)
    docker_version_dot_escaped="${DOCKER_VERSION//./\\.}"
    docker_version_dot_plus_escaped="${docker_version_dot_escaped//+/\\+}"
    # Regex needs to handle debian package version number format: https://www.systutorials.com/docs/linux/man/5-deb-version/
    docker_version_regex="^(.+:)?${docker_version_dot_plus_escaped}([\\.\\+ ~:-]|$)"
    set +e # Don't exit if finding version fails - will handle gracefully
    cli_version_suffix="=$(apt-cache madison ${cli_package_name} | awk -F"|" '{print $2}' | sed -e 's/^[ \t]*//' | grep -E -m 1 "${docker_version_regex}")"
    set -e
    if [ -z "${cli_version_suffix}" ] || [ "${cli_version_suffix}" = "=" ]; then
        echo "(!) No full or partial Docker / Moby version match found for \"${DOCKER_VERSION}\" on OS ${ID} ${VERSION_CODENAME} (${architecture}). Available versions:"
        apt-cache madison ${cli_package_name} | awk -F"|" '{print $2}' | grep -oP '^(.+:)?\K.+'
        exit 1
    fi
    echo "cli_version_suffix ${cli_version_suffix}"
fi

# Version matching for moby-buildx
if [ "${USE_MOBY}" = "true" ]; then
    if [ "${MOBY_BUILDX_VERSION}" = "latest" ]; then
        # Empty, meaning grab whatever "latest" is in apt repo
        buildx_version_suffix=""
    else
        buildx_version_dot_escaped="${MOBY_BUILDX_VERSION//./\\.}"
        buildx_version_dot_plus_escaped="${buildx_version_dot_escaped//+/\\+}"
        buildx_version_regex="^(.+:)?${buildx_version_dot_plus_escaped}([\\.\\+ ~:-]|$)"
        set +e
            buildx_version_suffix="=$(apt-cache madison moby-buildx | awk -F"|" '{print $2}' | sed -e 's/^[ \t]*//' | grep -E -m 1 "${buildx_version_regex}")"
        set -e
        if [ -z "${buildx_version_suffix}" ] || [ "${buildx_version_suffix}" = "=" ]; then
            err "No full or partial moby-buildx version match found for \"${MOBY_BUILDX_VERSION}\" on OS ${ID} ${VERSION_CODENAME} (${architecture}). Available versions:"
            apt-cache madison moby-buildx | awk -F"|" '{print $2}' | grep -oP '^(.+:)?\K.+'
            exit 1
        fi
        echo "buildx_version_suffix ${buildx_version_suffix}"
    fi
fi


docker_home="/usr/libexec/docker"
cli_plugins_dir="${docker_home}/cli-plugins"

install_compose_fallback(){
    local url=$1
    local repo_url=$(get_github_api_repo_url "$url")
    echo -e "\n(!) Failed to fetch the latest artifacts for docker-compose v${compose_version}..."
    get_previous_version "${url}" "${repo_url}" compose_version
    echo -e "\nAttempting to install v${compose_version}"
    curl -fsSL "https://github.com/docker/compose/releases/download/v${compose_version}/docker-compose-linux-${target_compose_arch}" -o ${docker_compose_path}
}

# Install Docker / Moby CLI if not already installed
if type docker > /dev/null 2>&1; then
    echo "Docker / Moby CLI already installed."
else
    if [ "${USE_MOBY}" = "true" ]; then
        buildx=()
        if [ "${INSTALL_DOCKER_BUILDX}" = "true" ]; then
            buildx=(moby-buildx${buildx_version_suffix})
        fi
        apt-get -y install --no-install-recommends ${cli_package_name}${cli_version_suffix} "${buildx[@]}" || { err "It seems packages for moby not available in OS ${ID} ${VERSION_CODENAME} (${architecture}). To resolve, either: (1) set feature option '\"moby\": false' , or (2) choose a compatible OS version (eg: 'ubuntu-24.04')." ; exit 1 ; }
        apt-get -y install --no-install-recommends moby-compose || echo "(*) Package moby-compose (Docker Compose v2) not available for OS ${ID} ${VERSION_CODENAME} (${architecture}). Skipping."
    else
        buildx=()
        if [ "${INSTALL_DOCKER_BUILDX}" = "true" ]; then
            buildx=(docker-buildx-plugin)
        fi
        apt-get -y install --no-install-recommends ${cli_package_name}${cli_version_suffix} "${buildx[@]}" docker-compose-plugin
        buildx_path="/usr/libexec/docker/cli-plugins/docker-buildx"
        # Older versions of Docker CE installs buildx as part of the CLI package
        if [ "${INSTALL_DOCKER_BUILDX}" = "false" ] && [ -f "${buildx_path}" ]; then
            echo "(*) Removing docker-buildx installed from docker-ce-cli since installDockerBuildx is disabled..."
            rm -f "${buildx_path}"
        fi
    fi
    unset buildx buildx_path
fi

# If 'docker-compose' command is to be included
if [ "${DOCKER_DASH_COMPOSE_VERSION}" != "none" ]; then
    case "${architecture}" in
        amd64) target_compose_arch=x86_64 ;;
        arm64) target_compose_arch=aarch64 ;;
        *)
            echo "(!) Docker outside of docker does not support machine architecture '$architecture'. Please use an x86-64 or ARM64 machine."
            exit 1
    esac
    docker_compose_path="/usr/local/bin/docker-compose"
    # Install Docker Compose if not already installed  and is on a supported architecture
    if type docker-compose > /dev/null 2>&1; then
        echo "Docker Compose already installed."
    elif [ "${DOCKER_DASH_COMPOSE_VERSION}" = "v1" ]; then
        err "The final Compose V1 release, version 1.29.2, was May 10, 2021. These packages haven't received any security updates since then. Use at your own risk."
        INSTALL_DOCKER_COMPOSE_SWITCH="false"

        if [ "${target_compose_arch}" = "x86_64" ]; then
            echo "(*) Installing docker compose v1..."
            curl -fsSL "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-Linux-x86_64" -o ${docker_compose_path}
            chmod +x ${docker_compose_path}

            # Download the SHA256 checksum
            DOCKER_COMPOSE_SHA256="$(curl -sSL "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-Linux-x86_64.sha256" | awk '{print $1}')"
            echo "${DOCKER_COMPOSE_SHA256}  ${docker_compose_path}" > docker-compose.sha256sum
            sha256sum -c docker-compose.sha256sum --ignore-missing
        elif [ "${VERSION_CODENAME}" = "bookworm" ]; then
            err "Docker compose v1 is unavailable for 'bookworm' on Arm64. Kindly switch to use v2"
            exit 1
        else
            # Use pip to get a version that runs on this architecture
            check_packages python3-minimal python3-pip libffi-dev python3-venv
            echo "(*) Installing docker compose v1 via pip..."
            export PYTHONUSERBASE=/usr/local
            pip3 install --disable-pip-version-check --no-cache-dir --user "Cython<3.0" pyyaml wheel docker-compose --no-build-isolation
        fi
    else
        compose_version=${DOCKER_DASH_COMPOSE_VERSION#v}
        docker_compose_url="https://github.com/docker/compose"
        find_version_from_git_tags compose_version "$docker_compose_url" "tags/v"
        echo "(*) Installing docker-compose ${compose_version}..."
        curl -fsSL "https://github.com/docker/compose/releases/download/v${compose_version}/docker-compose-linux-${target_compose_arch}" -o ${docker_compose_path} || {
            install_compose_fallback "$docker_compose_url" "$compose_version" "$target_compose_arch" "$docker_compose_path"
        }
        chmod +x ${docker_compose_path}

        # Download the SHA256 checksum
        DOCKER_COMPOSE_SHA256="$(curl -sSL "https://github.com/docker/compose/releases/download/v${compose_version}/docker-compose-linux-${target_compose_arch}.sha256" | awk '{print $1}')"
        echo "${DOCKER_COMPOSE_SHA256}  ${docker_compose_path}" > docker-compose.sha256sum
        sha256sum -c docker-compose.sha256sum --ignore-missing

        mkdir -p ${cli_plugins_dir}
        cp ${docker_compose_path} ${cli_plugins_dir}
    fi
fi

# Install docker-compose switch if not already installed - https://github.com/docker/compose-switch#manual-installation
if [ "${INSTALL_DOCKER_COMPOSE_SWITCH}" = "true" ] && ! type compose-switch > /dev/null 2>&1; then
    if type docker-compose > /dev/null 2>&1; then
        echo "(*) Installing compose-switch..."
        current_compose_path="$(which docker-compose)"
        target_compose_path="$(dirname "${current_compose_path}")/docker-compose-v1"
        compose_switch_version="latest"
        compose_switch_url="https://github.com/docker/compose-switch"
        find_version_from_git_tags compose_switch_version "${compose_switch_url}"
        curl -fsSL "https://github.com/docker/compose-switch/releases/download/v${compose_switch_version}/docker-compose-linux-${architecture}" -o /usr/local/bin/compose-switch || install_compose_switch_fallback "${compose_switch_url}"
        chmod +x /usr/local/bin/compose-switch
        # TODO: Verify checksum once available: https://github.com/docker/compose-switch/issues/11
        # Setup v1 CLI as alternative in addition to compose-switch (which maps to v2)
        mv "${current_compose_path}" "${target_compose_path}"
        update-alternatives --install ${docker_compose_path} docker-compose /usr/local/bin/compose-switch 99
        update-alternatives --install ${docker_compose_path} docker-compose "${target_compose_path}" 1
    else
        err "Skipping installation of compose-switch as docker compose is unavailable..."
    fi
fi

# Setup a docker group in the event the docker socket's group is not root
if ! grep -qE '^docker:' /etc/group; then
    echo "(*) Creating missing docker group..."
    groupadd --system docker
fi

# Remarking this out to restore functionality in Azure VMs.  ID 999 is a reserved group ID
# Ensure docker group gid is 999
# if [ "$(getent group docker | cut -d: -f3)" != "999" ]; then
#     echo "(*) Updating docker group gid to 999..."
#     groupmod -g 999 docker
# fi


usermod -aG docker "${USERNAME}"

# If init file already exists, exit
if [ -f "/usr/local/share/docker-init.sh" ]; then
    # Clean up
    rm -rf /var/lib/apt/lists/*
    exit 0
fi
echo "docker-init doesn't exist, adding..."

# By default, make the source and target sockets the same
if [ "${SOURCE_SOCKET}" != "${TARGET_SOCKET}" ]; then
    touch "${SOURCE_SOCKET}"
    ln -s "${SOURCE_SOCKET}" "${TARGET_SOCKET}"
fi

# Add a stub if not adding non-root user access, user is root
if [ "${ENABLE_NONROOT_DOCKER}" = "false" ] || [ "${USERNAME}" = "root" ]; then
    echo -e '#!/usr/bin/env bash\nexec "$@"' > /usr/local/share/docker-init.sh
    chmod +x /usr/local/share/docker-init.sh
    # Clean up
    rm -rf /var/lib/apt/lists/*
    exit 0
fi

DOCKER_GID="$(grep -oP '^docker:x:\K[^:]+' /etc/group)"

# If enabling non-root access and specified user is found, setup socat and add script
chown -h "${USERNAME}":root "${TARGET_SOCKET}"
check_packages socat
tee /usr/local/share/docker-init.sh > /dev/null \
<< EOF
#!/usr/bin/env bash
#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

set -e

SOCAT_PATH_BASE=/tmp/vscr-docker-from-docker
SOCAT_LOG=\${SOCAT_PATH_BASE}.log
SOCAT_PID=\${SOCAT_PATH_BASE}.pid

# Wrapper function to only use sudo if not already root
sudoIf()
{
    if [ "\$(id -u)" -ne 0 ]; then
        sudo "\$@"
    else
        "\$@"
    fi
}

# Log messages
log()
{
    echo -e "[\$(date)] \$@" | sudoIf tee -a \${SOCAT_LOG} > /dev/null
}

echo -e "\n** \$(date) **" | sudoIf tee -a \${SOCAT_LOG} > /dev/null
log "Ensuring ${USERNAME} has access to ${SOURCE_SOCKET} via ${TARGET_SOCKET}"

# If enabled, try to update the docker group with the right GID. If the group is root,
# fall back on using socat to forward the docker socket to another unix socket so
# that we can set permissions on it without affecting the host.
if [ "${ENABLE_NONROOT_DOCKER}" = "true" ] && [ "${SOURCE_SOCKET}" != "${TARGET_SOCKET}" ] && [ "${USERNAME}" != "root" ] && [ "${USERNAME}" != "0" ]; then
    SOCKET_GID=\$(stat -c '%g' ${SOURCE_SOCKET})
    if [ "\${SOCKET_GID}" != "0" ] && [ "\${SOCKET_GID}" != "${DOCKER_GID}" ] && ! grep -E ".+:x:\${SOCKET_GID}" /etc/group; then
        sudoIf groupmod --gid "\${SOCKET_GID}" docker
    else
        # Enable proxy if not already running
        if [ ! -f "\${SOCAT_PID}" ] || ! ps -p \$(cat \${SOCAT_PID}) > /dev/null; then
            log "Enabling socket proxy."
            log "Proxying ${SOURCE_SOCKET} to ${TARGET_SOCKET} for vscode"
            sudoIf rm -rf ${TARGET_SOCKET}
            (sudoIf socat UNIX-LISTEN:${TARGET_SOCKET},fork,mode=660,user=${USERNAME},backlog=128 UNIX-CONNECT:${SOURCE_SOCKET} 2>&1 | sudoIf tee -a \${SOCAT_LOG} > /dev/null & echo "\$!" | sudoIf tee \${SOCAT_PID} > /dev/null)
        else
            log "Socket proxy already running."
        fi
    fi
    log "Success"
fi

# Execute whatever commands were passed in (if any). This allows us
# to set this script to ENTRYPOINT while still executing the default CMD.
set +e
exec "\$@"
EOF
chmod +x /usr/local/share/docker-init.sh
chown ${USERNAME}:root /usr/local/share/docker-init.sh

# Clean up
rm -rf /var/lib/apt/lists/*

echo "Done!"
