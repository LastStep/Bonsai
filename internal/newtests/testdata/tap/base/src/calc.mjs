// A fixture of the new-tests check: sub is wrong at the base and right on the branch.
export const add = (a, b) => a + b;
// At the base sub adds, which only a test with a non-zero b sees.
export const sub = (a, b) => a + b;
