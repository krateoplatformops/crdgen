package coder

import (
	"errors"
	"os"
	"path/filepath"
)

type Resource struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}

type Options struct {
	Module  string
	Workdir string
}

func Do(res *Resource, cfg Options) error {
	err := CreateGenerateDotGo(cfg.Workdir)
	if err != nil {
		return err
	}

	err = CreateTypesDotGo(cfg.Workdir, res)
	if err != nil {
		return err
	}

	err = CreateGroupVersionInfoDotGo(cfg.Workdir, res)
	if err != nil {
		return err
	}

	err = CreateApisDotGo(res, cfg)
	if err != nil {
		return err
	}

	if res.Managed {
		err := GenerateManaged(cfg.Workdir, res)
		if err != nil {
			return err
		}

		err = GenerateManagedList(cfg.Workdir, res)
		if err != nil {
			return err
		}
	}

	err = os.Mkdir(filepath.Join(cfg.Workdir, "crds"), os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	err = os.Mkdir(filepath.Join(cfg.Workdir, "hack"), os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	fp, err := os.Create(filepath.Join(cfg.Workdir, "hack", "boilerplate.go.txt"))
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = fp.WriteString("// Copyright 2024 Krateo SRL.")
	return err
}
