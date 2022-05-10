package generator

import (
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util/ctx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

const (
	wd       = "wd"
	etc      = "etc"
	internal = "internal"
	config   = "config"
	logic    = "logic"
	server   = "server"
	svc      = "svc"
	pb       = "pb"
	protoGo  = "proto-go"
	call     = "call"
	calls    = "calls"
)

type (
	// DirContext defines a rpc service directories context
	DirContext interface {
		GetCall() Dir
		_GetCall(serviceName string) Dir
		GetEtc() Dir
		GetInternal() Dir
		GetConfig() Dir
		GetLogic() Dir
		_GetLogic(serviceName string) Dir
		GetServer() Dir
		GetSvc() Dir
		GetPb() Dir
		GetProtoGo() Dir
		GetMain() Dir
		GetServiceName() stringx.String
		SetPbDir(pbDir, grpcDir string)
	}

	// Dir defines a directory
	Dir struct {
		Base     string
		Filename string
		Package  string
	}

	defaultDirContext struct {
		inner       map[string]Dir
		calls       map[string]Dir
		logics      map[string]Dir
		serviceName stringx.String // service name is goctl-specific
		ctx         *ctx.ProjectContext
	}
)

func mkdir(ctx *ctx.ProjectContext, proto parser.Proto, generator *Generator, zctx *ZRpcContext) (DirContext, error) {
	inner := make(map[string]Dir)
	calls := make(map[string]Dir)
	logics := make(map[string]Dir)
	etcDir := filepath.Join(ctx.WorkDir, "etc")
	internalDir := filepath.Join(ctx.WorkDir, "internal")
	configDir := filepath.Join(internalDir, "config")
	logicDir := filepath.Join(internalDir, "logic")
	serverDir := filepath.Join(internalDir, "server")
	svcDir := filepath.Join(internalDir, "svc")
	pbDir := filepath.Join(ctx.WorkDir, proto.GoPackage)
	protoGoDir := pbDir
	if zctx != nil {
		pbDir = zctx.ProtoGenGrpcDir
		protoGoDir = zctx.ProtoGenGoDir
	}

	for _, service := range proto.Services {
		callDir := filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(service.Name).ToCamel()))
		if strings.EqualFold(service.Name, proto.GoPackage) {
			callDir = filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(service.Name+"_client").ToCamel()))
		}

		calls[service.Name] = Dir{
			Filename: callDir,
			Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(callDir, ctx.Dir))),
			Base:     filepath.Base(callDir),
		}

		logicDir := filepath.Join(internalDir, "logic", strings.ToLower(service.Name))
		logics[service.Name] = Dir{
			Filename: logicDir,
			Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(logicDir, ctx.Dir))),
			Base:     filepath.Base(logicDir),
		}
	}

	callDir := filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(proto.Service.Name).ToCamel()))
	if strings.EqualFold(proto.Service.Name, proto.GoPackage) {
		callDir = filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(proto.Service.Name+"_client").ToCamel()))
	}

	inner[wd] = Dir{
		Filename: ctx.WorkDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(ctx.WorkDir, ctx.Dir))),
		Base:     filepath.Base(ctx.WorkDir),
	}
	inner[etc] = Dir{
		Filename: etcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(etcDir, ctx.Dir))),
		Base:     filepath.Base(etcDir),
	}
	inner[internal] = Dir{
		Filename: internalDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(internalDir, ctx.Dir))),
		Base:     filepath.Base(internalDir),
	}
	inner[config] = Dir{
		Filename: configDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(configDir, ctx.Dir))),
		Base:     filepath.Base(configDir),
	}
	inner[logic] = Dir{
		Filename: logicDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(logicDir, ctx.Dir))),
		Base:     filepath.Base(logicDir),
	}
	inner[server] = Dir{
		Filename: serverDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(serverDir, ctx.Dir))),
		Base:     filepath.Base(serverDir),
	}
	inner[svc] = Dir{
		Filename: svcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(svcDir, ctx.Dir))),
		Base:     filepath.Base(svcDir),
	}

	inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(pbDir, ctx.Dir))),
		Base:     filepath.Base(pbDir),
	}

	inner[protoGo] = Dir{
		Filename: protoGoDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(protoGoDir, ctx.Dir))),
		Base:     filepath.Base(protoGoDir),
	}

	inner[call] = Dir{
		Filename: callDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(callDir, ctx.Dir))),
		Base:     filepath.Base(callDir),
	}

	dirMaps := []map[string]Dir{inner}
	if generator.multiServiceEnabled {
		dirMaps = append(dirMaps, calls, logics)
	}
	for _, x := range dirMaps {
		for _, v := range x {
			err := pathx.MkdirIfNotExist(v.Filename)
			if err != nil {
				return nil, err
			}
		}
	}
	serviceName := strings.TrimSuffix(proto.Name, filepath.Ext(proto.Name))
	return &defaultDirContext{
		ctx:         ctx,
		inner:       inner,
		calls:       calls,
		logics:      logics,
		serviceName: stringx.From(strings.ReplaceAll(serviceName, "-", "")),
	}, nil
}

func (d *defaultDirContext) SetPbDir(pbDir, grpcDir string) {
	d.inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(d.ctx.Path, strings.TrimPrefix(pbDir, d.ctx.Dir))),
		Base:     filepath.Base(pbDir),
	}

	d.inner[protoGo] = Dir{
		Filename: grpcDir,
		Package:  filepath.ToSlash(filepath.Join(d.ctx.Path, strings.TrimPrefix(grpcDir, d.ctx.Dir))),
		Base:     filepath.Base(grpcDir),
	}
}

func (d *defaultDirContext) GetCall() Dir {
	return d.inner[call]
}

func (d *defaultDirContext) GetEtc() Dir {
	return d.inner[etc]
}

func (d *defaultDirContext) GetInternal() Dir {
	return d.inner[internal]
}

func (d *defaultDirContext) GetConfig() Dir {
	return d.inner[config]
}

func (d *defaultDirContext) GetLogic() Dir {
	return d.inner[logic]
}

func (d *defaultDirContext) GetServer() Dir {
	return d.inner[server]
}

func (d *defaultDirContext) GetSvc() Dir {
	return d.inner[svc]
}

func (d *defaultDirContext) GetPb() Dir {
	return d.inner[pb]
}

func (d *defaultDirContext) GetProtoGo() Dir {
	return d.inner[protoGo]
}

func (d *defaultDirContext) GetMain() Dir {
	return d.inner[wd]
}

func (d *defaultDirContext) GetServiceName() stringx.String {
	return d.serviceName
}

func (d *defaultDirContext) _GetCall(serviceName string) Dir {
	return d.calls[serviceName]
}

func (d *defaultDirContext) _GetLogic(serviceName string) Dir {
	return d.logics[serviceName]
}

// Valid returns true if the directory is valid
func (d *Dir) Valid() bool {
	return len(d.Filename) > 0 && len(d.Package) > 0
}
