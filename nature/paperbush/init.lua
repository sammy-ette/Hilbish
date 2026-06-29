local inspect = require 'inspect'
local lexer = require 'nature.paperbush.lexer'
local classifier = require 'nature.paperbush.classifier'
local codegen = require 'nature.paperbush.codegen'
local runtime = require 'nature.paperbush.runtime'
local diagnostic = require 'nature.paperbush.diagnostic'
local luaerr = require 'nature.paperbush.luaerr'

hilbish.runner.add('paperbush', {
    run = function(input)
        local code = codegen.chunk(input)
        local diags = codegen.takeDiagnostics()
        local map = codegen.lastMap()

        ---@diagnostic disable-next-line: undefined-global
        if paperbushDebug then
            print(code)
            for _, d in ipairs(diags) do print(d.severity, d.code, d.message) end
        end

        local hasError = false
        for _, d in ipairs(diags) do
            diagnostic.report(d, input)
            if d.severity == 'error' then hasError = true end
        end
        if hasError then
            return { input = input, exitCode = 125 }
        end

        local f, err = load(code, '=paperbush')
        if not f then
            diagnostic.report(luaerr.fromLoad(err, input, map), input)
            return { input = input, exitCode = 125 }
        end

        runtime.lastExit, runtime.lastErr = 0, nil
        local ok, res = pcall(f)
        if not ok then
            diagnostic.report(luaerr.fromRuntime(res, input, map), input)
            return { input = input, exitCode = 126 }
        end

        if res then
            print(inspect(res))
        end

        return {
            input = input,
            exitCode = runtime.lastExit or 0,
            err = runtime.lastErr
        }
    end,
    validate = function(input)
        local toks = lexer.tokenize(input)
        for _, tok in ipairs(toks) do
            if tok.incomplete then return false end
        end

        local ok, _, incomplete = pcall(classifier.parse, input)
        if ok and incomplete then return false end

        return true
    end
})
