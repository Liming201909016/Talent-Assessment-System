package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/talent-assessment/refactored/pkg/libreofficepdf"
)

type managementTraitsReportDiagnostic struct {
	stage string
	class string
}

func (d *managementTraitsReportDiagnostic) failure(err error) {
	if d != nil && d.class == "" {
		if err == nil {
			err = ErrManagementTraitsRuntimeInvalid
		}
		d.class = managementTraitsReportErrorClass(err)
	}
}

func (d *managementTraitsReportDiagnostic) emit() {
	if d == nil || d.class == "" {
		return
	}
	switch d.stage {
	case "input", "schema", "load", "marshal", "render", "write", "reload", "revision", "persist":
		slog.Error("management_traits_report_failure", "stage", d.stage, "class", d.class)
	}
}

func managementTraitsReportErrorClass(err error) string {
	var conversion *libreofficepdf.ConversionError
	if errors.As(err, &conversion) {
		class := conversion.DiagnosticClass()
		switch class {
		case "lo_config", "lo_input_size", "lo_input_name", "lo_queue", "lo_workspace", "lo_workspace_permission", "lo_input_write", "lo_command", "lo_output_open", "lo_output_read", "lo_output_invalid":
			cause := managementTraitsReportCauseClass(conversion.Unwrap())
			if cause != "unknown" {
				return class + "_" + cause
			}
			return class
		}
	}
	return managementTraitsReportCauseClass(err)
}

func managementTraitsReportCauseClass(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "context_deadline"
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case errors.Is(err, os.ErrPermission):
		return "os_permission"
	}
	var path *os.PathError
	if errors.As(err, &path) {
		return "os_path"
	}
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "os_not_exist"
	case errors.Is(err, os.ErrExist):
		return "os_exists"
	case errors.Is(err, ErrManagementTraitsRuntimeInvalid):
		return "validation"
	case errors.Is(err, ErrManagementTraitsRuntimeClosed):
		return "runtime_closed"
	}
	var database *mysql.MySQLError
	if errors.As(err, &database) {
		return "mysql_" + strconv.FormatUint(uint64(database.Number), 10)
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return "exec_exit"
	}
	var command *exec.Error
	if errors.As(err, &command) {
		return "exec_start"
	}
	return "unknown"
}
