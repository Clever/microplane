# !usr/bin/python3
import os
import json
import subprocess


def parseAndInsert(path, after, insert):
    with open(path, "r") as f:
        text = f.read()
    index = text.find(after)
    if index == -1:
        return
    idx = index + len(after)
    updated = text[:idx] + insert + text[idx:]
    with open(path, "w") as f:
        f.write(updated)


def installPackage(cwd, packageName):
    with open(os.path.join(cwd, "package.json"), "r") as fpj:
        pj = json.load(fpj)
    if packageName in pj["dependencies"] or packageName in pj["devDependencies"]:
        print(f"{packageName} is installed ready")
        return
    print(subprocess.check_output(["npm", "i", packageName]).decode())


def updateEslintrc(cwd, ruleName):
    # infer .js or .yml
    if os.path.exists(os.path.join(cwd, ".eslintrc.js")):
        eslintrc = ".eslintrc.js"
        path = os.path.join(cwd, eslintrc)
        # insert plugin as well as rule
        parseAndInsert(path, "plugins: [", f'\n    "@clever",')
        parseAndInsert(path, "rules: {", f'\n    "{ruleName}": "error",')
    elif os.path.exists(os.path.join(cwd, ".eslintrc.yml")):
        eslintrc = ".eslintrc.yml"
        path = os.path.join(cwd, eslintrc)
        parseAndInsert(path, "plugins:", f'\n  - "@clever"')
        parseAndInsert(path, "rules:", f'\n  "{ruleName}": "error"')
    print(f"Updated {eslintrc} to have plugin and rule")
    return


def runEslintFix(cwd, path):
    print(
        subprocess.check_output(
            [
                f"{cwd}/node_modules/.bin/eslint",
                "--ext",
                "ts",
                "--fix",
                "--ignore-pattern",
                "*/node_modules/*",
                path,
            ],
        ).decode()
    )

def find_app_listen(cwd):
    return subprocess.check_output(
        ["rg", "app.listen", "-l"]
    ).decode()

def main():
    cwd = os.getcwd()
    print(f"CWD = {cwd}")
    # install eslint if needed
    installPackage(cwd, "eslint")
    # install @clever/eslint-plugin
    installPackage(cwd, "@clever/eslint-plugin@latest")
    # run npm install
    print(subprocess.check_output(["npm", "i"]).decode())

    # update eslintrc if needed
    updateEslintrc(cwd, "@clever/no-app-listen-without-localhost")
    # run eslint --fix on app.listen file
    runEslintFix(cwd, find_app_listen(cwd))

if __name__ == "__main__":
    main()
