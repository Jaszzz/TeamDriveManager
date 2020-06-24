package generate

import (
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/fionera/TeamDriveManager/api"
	. "github.com/fionera/TeamDriveManager/config"
)

func NewGenerateRcloneCommand() cli.Command {
	return cli.Command{
		Name:   "rclone",
		Usage:  "Generate a rclone config",
		Action: CmdGenerateRclone,
		Flags:  []cli.Flag{},
	}
}

func CmdGenerateRclone(c *cli.Context) {
	filter := strings.Join(c.Args(), " ")

	if filter != "" {
		logrus.Infof("Using filter `%s`", filter)
	}

	tokenSource, err := api.NewTokenSource(App.AppConfig.ServiceAccountFile, App.AppConfig.Impersonate)
	if err != nil {
		logrus.Error(err)
		return
	}

	driveApi, err := api.NewDriveService(tokenSource)
	if err != nil {
		logrus.Error(err)
		return
	}

	boolResponse := false
	confirm := &survey.Confirm{
		Message: "Use Domain Admin access?",
		Default: false,
	}

	err = survey.AskOne(confirm, &boolResponse, nil)
	if err != nil {
		logrus.Panic(err)
		return
	}

	var list = api.ListTeamDrives
	if boolResponse {
		list = api.ListAllTeamDrives
	}

	teamDrives, err := list(driveApi)
	if err != nil {
		logrus.Panic(err)
		return
	}

	sb := strings.Builder{}
	for _, teamDrive := range teamDrives {
		if !strings.HasPrefix(teamDrive.Name, filter) {
			continue
		}

		name := strings.Map(func(r rune) rune {
			if r > unicode.MaxASCII {
				return -1
			}
			return r
		}, teamDrive.Name)

		name = strings.Map(func(r rune) rune {
			if r == '/' ||
				r == '_' ||
				r == '.' {
				return '-'
			}
			return r
		}, name)

		sb.WriteString(fmt.Sprintf("[%s]\n", name))
		sb.WriteString("type = drive\n")
		sb.WriteString("scope = drive\n")
		sb.WriteString(fmt.Sprintf("team_drive = %s\n", teamDrive.Id))
		sb.WriteString("\n")
	}

	fmt.Println(sb.String())
	_ = ioutil.WriteFile("rclone.conf", []byte(sb.String()), 0644)
}
