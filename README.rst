nazuna
======

nazuna is a layered dotfiles management tool.

.. image:: https://semaphoreci.com/api/v1/hattya/nazuna/branches/master/badge.svg
   :target: https://semaphoreci.com/hattya/nazuna

.. image:: https://ci.appveyor.com/api/projects/status/2eg4vbro37mhsdk0/branch/master?svg=true
   :target: https://ci.appveyor.com/project/hattya/nazuna

.. image:: https://codecov.io/gh/hattya/nazuna/branch/master/graph/badge.svg
   :target: https://codecov.io/gh/hattya/nazuna


Install
-------

.. code:: console

   $ go get -u github.com/hattya/nazuna/cmd/nzn


Usage
-----

.. code:: console

   $ nzn init --vcs git
   $ nzn layer -c master
   $ cp .gitconfig .nzn/r/master
   $ nzn vcs add .
   $ nzn vcs commit -m "Initial import"
   [master (root-commit) 1234567] Initial import
    2 files changed, 8 insertions(+)
    create mode 100644 master/.gitconfig
    create mode 100644 nazuna.json
   $ rm .gitconfig
   $ nzn update
   link .gitconfig --> master
   1 updated, 0 removed, 0 failed
   $ readlink .gitconfig
   .nzn/r/master/.gitconfig


License
-------

nazuna is distributed under the terms of the MIT License.
