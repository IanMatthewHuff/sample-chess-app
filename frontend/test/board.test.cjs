const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

test('rendering a new position retains square elements', () => {
    const boardFile = fs.readFileSync(path.join(__dirname, '..', 'src', 'board.ts'), 'utf8');
    const code = ts.transpileModule(boardFile, {
        compilerOptions: { module: ts.ModuleKind.CommonJS }
    }).outputText;
    const container = {
        children: [],
        replaceChildren(fragment) { this.children = fragment.children; }
    };
    const document = {
        getElementById: () => container,
        createDocumentFragment: () => ({
            children: [],
            appendChild(child) { this.children.push(child); }
        }),
        createElement: () => ({
            dataset: {},
            className: '',
            textContent: '',
            setAttribute(name, value) { this[name] = value; }
        })
    };
    const exports = {};
    vm.runInNewContext(code, {
        exports,
        document,
        require: () => ({ getPieceSymbol: piece => piece.type[0] })
    });
    const board = Array.from({ length: 8 }, () => Array(8).fill(null));
    board[6][4] = { type: 'pawn', color: 'white', hasMoved: false };
    exports.renderBoard(board, 'board');
    assert.equal(container.children.length, 64);
    const e2 = container.children[52];
    const e4 = container.children[36];
    assert.equal(e2.textContent, 'p');

    board[4][4] = board[6][4];
    board[6][4] = null;
    exports.renderBoard(board, 'board');
    assert.equal(container.children[52], e2);
    assert.equal(container.children[36], e4);
    assert.equal(e2.textContent, '');
    assert.equal(e4.textContent, 'p');

    exports.renderBoard(board, 'board', true);
    assert.equal(container.children[52], e2);
    assert.equal(container.children[27].textContent, 'p');
});
