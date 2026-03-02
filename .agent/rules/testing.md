# Go Testing Rules & Best Practices

When writing or modifying tests, you MUST adhere to the following rules:

1. **Test Location**: Put tests in the same directory as the code they test. Use the `_test` suffix for the package name if you are testing the exported API of the package (e.g., `package mypkg_test`), but keep it in the same package (e.g., `package mypkg`) if you need to test unexported internal methods.
2. **Table-Driven Tests**: ALWAYS use table-driven tests for multiple test cases. Structure your test cases cleanly with a `name` field for easy identification and use `t.Run()` to execute subtests.
3. **Mocks with Mockery**: Generate mocks using `vektra/mockery`. Mocks should be tightly scoped to the interfaces defined in your Domain/Ports layer (Hexagonal Architecture). Do NOT write mocks by hand unless absolutely necessary.
4. **Assertions**: Utilize `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` for clean and readable assertions instead of standard library `if err != nil { t.Errorf(...) }` boilerplate.
5. **Setup and Teardown**: Use `t.Cleanup()` for any teardown logic instead of `defer`, as it ensures proper cleanup even if a subtest fails or panics.
6. **Error Testing**: When testing for errors, use `require.ErrorIs` or `require.ErrorAs` for specific error assertions rather than comparing error strings.
7. **Race Detection**: All unit test suites should pass cleanly with the `-race` flag enabled. Avoid shared global state that could cause data races during tests.
