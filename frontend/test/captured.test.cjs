const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

function loadModule(file, document, requireModule = () => ({})) {
    const source = fs.readFileSync(path.join(__dirname, '..', 'src', file), 'utf8');
    const code = ts.transpileModule(source, {
        compilerOptions: { module: ts.ModuleKind.CommonJS }
    }).outputText;
    const exports = {};
    vm.runInNewContext(code, { exports, document, require: requireModule });
    return exports;
}

test('captured pieces use the same symbols as the board', () => {
    const makeElement = () => ({
        dataset: {},
        children: [],
        textContent: '',
        replaceChildren(...children) { this.children = children; },
        setAttribute(name, value) { this[name] = value; }
    });
    const white = makeElement();
    const black = makeElement();
    const document = {
        getElementById: id => id === 'captured-by-white' ? white : black,
        createElement: makeElement
    };
    const pieces = loadModule('pieces.ts', document);
    const { renderCapturedPieces } = loadModule('captured.ts', document, () => pieces);
    const pawn = color => ({ type: 'pawn', color, hasMoved: true });
    const moves = [
        { piece: pawn('white'), captured: pawn('black') },
        { piece: pawn('black'), captured: pawn('white') }
    ];
    renderCapturedPieces(moves);
    assert.equal(white.children[0].textContent, '♟');
    assert.equal(white.children[0].className, 'captured-piece black');
    assert.equal(white.children[0]['aria-label'], 'black pawn');
    assert.equal(black.children[0].textContent, '♙');
    assert.equal(black.children[0].className, 'captured-piece white');
    assert.equal(black.children[0]['aria-label'], 'white pawn');

    const chip = white.children[0];
    renderCapturedPieces(moves);
    assert.equal(white.children[0], chip);
    renderCapturedPieces([]);
    assert.equal(white.textContent, 'None');
    assert.equal(black.textContent, 'None');
});
