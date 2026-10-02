const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

test('legal captures use a red ring instead of a green dot, even for en passant', () => {
    const source = fs.readFileSync(path.join(__dirname, '..', 'src', 'highlights.ts'), 'utf8');
    const code = ts.transpileModule(source, {
        compilerOptions: { module: ts.ModuleKind.CommonJS }
    }).outputText;
    const exports = {};
    vm.runInNewContext(code, { exports });
    const classes = new Set();
    const attributes = new Map();
    const square = {
        classList: {
            toggle(name, active) {
                if (active) classes.add(name);
                else classes.delete(name);
            }
        },
        getAttribute: name => attributes.get(name),
        setAttribute: (name, value) => attributes.set(name, value)
    };
    const lastMove = { from: { row: 3, col: 4 }, to: { row: 2, col: 3 } };
    const capture = { row: 2, col: 3, isCapture: true };
    exports.highlightSquare(square, capture, null, lastMove.from, [capture], lastMove);
    assert.ok(classes.has('capture-move'));
    assert.ok(classes.has('last-move-to'));
    assert.ok(!classes.has('valid-move'));
    assert.equal(attributes.get('aria-label'), 'd6, capture available');

    exports.highlightSquare(square, capture, null, null,
        [{ row: 2, col: 3, isCapture: false }]);
    assert.ok(classes.has('valid-move'));
    assert.ok(!classes.has('capture-move'));
    assert.ok(!classes.has('last-move-to'));
    assert.equal(attributes.get('aria-label'), 'd6, legal move');

    const king = { type: 'king', color: 'white', hasMoved: true };
    exports.highlightSquare(square, { row: 7, col: 4 }, king, null, [], undefined, 'white');
    assert.ok(classes.has('king-in-check'));
    assert.equal(attributes.get('aria-label'), 'e1 white king, in check');
    exports.highlightSquare(square, { row: 7, col: 4 }, king, null, [], undefined, null);
    assert.ok(!classes.has('king-in-check'));
    exports.highlightSquare(square, { row: 0, col: 4 },
        { ...king, color: 'black' }, null, [], undefined, 'black');
    assert.ok(classes.has('king-in-check'));

    const css = fs.readFileSync(path.join(__dirname, '..', 'src', 'styles.css'), 'utf8');
    assert.match(css, /\.square\.valid-move::after,\s*\.square\.capture-move::after\s*\{\s*content:\s*''/);
    assert.match(css, /\.square\.capture-move::after\s*\{[^}]*border:\s*4px solid #c73647/s);
    assert.match(css, /\.square\.king-in-check\s*\{[^}]*background-color:\s*#ee7777/s);
});
