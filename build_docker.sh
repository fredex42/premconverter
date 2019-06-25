#!/usr/bin/env bash -e

function abspath() {
    # generate absolute path from relative path
    # $1     : relative filename
    # return : absolute path
    if [ -d "$1" ]; then
        # dir
        (cd "$1"; pwd)
    elif [ -f "$1" ]; then
        # file
        if [[ $1 == */* ]]; then
            echo "$(cd "${1%/*}"; pwd)/${1##*/}"
        else
            echo "$(pwd)/$1"
        fi
    fi
}

BASEPATH=$(abspath "${BASH_SOURCE%/*}")

if [ "${CIRCLE_BUILD}" == "" ]; then
    declare -x CIRCLE_BUILD=DEV
fi

cd ${BASEPATH}; make
cd ${BASEPATH}; docker build . -t andyg42/premconverter:${CIRCLE_BUILD}
