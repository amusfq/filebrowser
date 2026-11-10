import * as monaco from "monaco-editor/editor/editor.api.js";
import EditorWorker from "monaco-editor/editor/editor.worker?worker";
import "monaco-editor/editor/contrib/find/browser/findController.js";
import "monaco-editor/editor/contrib/clipboard/browser/clipboard.js";
import "monaco-editor/editor/contrib/contextmenu/browser/contextmenu.js";
import "monaco-editor/editor/contrib/folding/browser/folding.js";
import "monaco-editor/editor/contrib/bracketMatching/browser/bracketMatching.js";
import "monaco-editor/editor/contrib/suggest/browser/suggestController.js";
import "monaco-editor/editor/contrib/wordOperations/browser/wordOperations.js";
import "monaco-editor/editor/contrib/linesOperations/browser/linesOperations.js";
import "monaco-editor/editor/contrib/tokenization/browser/tokenization.js";
import "monaco-editor/editor/standalone/browser/quickAccess/standaloneCommandsQuickAccess.js";
import "monaco-editor/editor/standalone/browser/quickAccess/standaloneGotoLineQuickAccess.js";
import { createHighlighterCore } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import { shikiToMonaco } from "@shikijs/monaco";
import { editorLanguages } from "./editor";
import javascriptGrammar from "shiki/langs/javascript.mjs";
import jsxGrammar from "shiki/langs/jsx.mjs";
import typescriptGrammar from "shiki/langs/typescript.mjs";
import tsxGrammar from "shiki/langs/tsx.mjs";
import jsonGrammar from "shiki/langs/json.mjs";
import htmlGrammar from "shiki/langs/html.mjs";
import cssGrammar from "shiki/langs/css.mjs";
import scssGrammar from "shiki/langs/scss.mjs";
import markdownGrammar from "shiki/langs/markdown.mjs";
import yamlGrammar from "shiki/langs/yaml.mjs";
import bashGrammar from "shiki/langs/bash.mjs";
import pythonGrammar from "shiki/langs/python.mjs";
import goGrammar from "shiki/langs/go.mjs";
import rustGrammar from "shiki/langs/rust.mjs";
import sqlGrammar from "shiki/langs/sql.mjs";
import vueGrammar from "shiki/langs/vue.mjs";
import svelteGrammar from "shiki/langs/svelte.mjs";
import phpGrammar from "shiki/langs/php.mjs";
import javaGrammar from "shiki/langs/java.mjs";
import cGrammar from "shiki/langs/c.mjs";
import cppGrammar from "shiki/langs/cpp.mjs";
import csharpGrammar from "shiki/langs/csharp.mjs";
import rubyGrammar from "shiki/langs/ruby.mjs";
import tomlGrammar from "shiki/langs/toml.mjs";
import iniGrammar from "shiki/langs/ini.mjs";
import dockerfileGrammar from "shiki/langs/dockerfile.mjs";
import xmlGrammar from "shiki/langs/xml.mjs";
import diffGrammar from "shiki/langs/diff.mjs";
import githubLight from "shiki/themes/github-light.mjs";
import githubDark from "shiki/themes/github-dark.mjs";
import monokai from "shiki/themes/monokai.mjs";
import solarizedLight from "shiki/themes/solarized-light.mjs";
import solarizedDark from "shiki/themes/solarized-dark.mjs";

self.MonacoEnvironment = { getWorker: () => new EditorWorker() };

// Initialize once: Shiki installs TextMate tokenizers and Monaco theme hooks globally.
let initialization: Promise<void> | undefined;
export function prepareMonaco() {
  return (initialization ||= createHighlighterCore({
    engine: createOnigurumaEngine(import("shiki/wasm")),
    themes: [githubLight, githubDark, monokai, solarizedLight, solarizedDark],
    langs: [
      javascriptGrammar,
      jsxGrammar,
      typescriptGrammar,
      tsxGrammar,
      jsonGrammar,
      htmlGrammar,
      cssGrammar,
      scssGrammar,
      markdownGrammar,
      yamlGrammar,
      bashGrammar,
      pythonGrammar,
      goGrammar,
      rustGrammar,
      sqlGrammar,
      vueGrammar,
      svelteGrammar,
      phpGrammar,
      javaGrammar,
      cGrammar,
      cppGrammar,
      csharpGrammar,
      rubyGrammar,
      tomlGrammar,
      iniGrammar,
      dockerfileGrammar,
      xmlGrammar,
      diffGrammar,
    ],
  })
    .then((highlighter) => {
      for (const [id, extensions] of Object.entries(editorLanguages)) {
        monaco.languages.register({
          id,
          extensions: extensions.map((ext) => `.${ext}`),
        });
      }
      shikiToMonaco(highlighter, monaco);
    })
    .catch((error) => {
      initialization = undefined;
      throw error;
    }));
}

export { monaco };
