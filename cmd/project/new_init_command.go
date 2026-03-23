package main

import appcreate "github.com/JackDrogon/project/internal/app/create"

func resolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	return appcreate.ResolveNewProjectArgs(lang, module, arg)
}

func projectNameFromGoModulePath(modulePath string) string {
	return appcreate.ProjectNameFromGoModulePath(modulePath)
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	return appcreate.ProjectNameFromTargetDir(targetDir)
}
