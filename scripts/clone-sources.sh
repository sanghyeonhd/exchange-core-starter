#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-$HOME/exchange-lab/sources}"
mkdir -p "$ROOT"
cd "$ROOT"

clone_or_pull() {
  local url="$1"
  local dir="$2"
  if [ -d "$dir/.git" ]; then
    echo "[pull] $dir"
    git -C "$dir" pull --ff-only || true
  else
    echo "[clone] $url -> $dir"
    git clone "$url" "$dir"
  fi
}

clone_or_pull https://github.com/opexdev/core.git opex-core
clone_or_pull https://gitee.com/source/coinglobalvip.git coinglobalvip
clone_or_pull https://gitee.com/coincoin/crypto-exchange.git coincoin-crypto-exchange
clone_or_pull https://github.com/Johnny1110/frizo-exchange.git frizo-exchange
clone_or_pull https://github.com/Johnny1110/futures_engine.git futures-engine
clone_or_pull https://github.com/0xae/omx-engine.git omx-engine
clone_or_pull https://github.com/Polygant/OpenCEX.git opencex
clone_or_pull https://github.com/Billy-Artifice/CoinExchange.git billy-coinexchange
clone_or_pull https://github.com/exchange-core/exchange-core.git exchange-core

