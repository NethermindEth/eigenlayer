# internal Package

The `internal` package contains code that will be used by packages under the `pkg` directory as well as by other packages under the `internal` directory. Users should not use this package directly. All packages under the `internal` directory should not contain any business logic.

We expect to have two main types of code under this directory:

1.  Simple common code that is used by multiple packages.

    These codes should be in a file inside the `common` directory. To understand what to put inside the `common` directory, please consult the [common package doc](./common/README.md).

2. Interfaces that are independent of the business domain and deserve their own packages. Examples include logging, configuration, and Docker interaction. In this case, you can create a new package at the root of this package to contain them.
