// A fixture of package testout's tests (internal/testout/testdata/tap). out.tap is what
//   node --test a.test.mjs test/b.test.mjs broken.test.mjs
// printed here with node 22 (its home-folder paths since made neutral). The run test runs single tests of this file
// and of test/b.test.mjs by the commands package testout gives.
import { describe, it, test } from 'node:test';
import assert from 'node:assert';
describe('outer suite', () => {
  it('inner one', () => { assert.strictEqual(1, 2); });
  it('inner two', () => {});
  describe('deeper', () => { it('leaf # hash', () => { throw new Error('boom\nsecond line'); }); });
});
test('top level', async (t) => {
  await t.test('sub a', () => {});
  await t.test('sub b', () => { assert.ok(false, 'nope'); });
});
test('skipped', { skip: 'real-checkouts-only: needs a real checkout' }, () => {});
test('todo one', { todo: 'later' }, () => {});
test('plain skip', (t) => { t.skip('no reason mark'); });
