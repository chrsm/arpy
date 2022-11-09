package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"bits.chrsm.org/arpy"
)

func main() {
	root.Execute()
}

var (
	root = &cobra.Command{
		Use:   "arpy",
		Short: "rpa pack and unpack tool",
		Long:  "rpa pack and unpack tool",
	}

	pack = &cobra.Command{
		Use:   "pack",
		Short: "pack an RPA",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			if out == "" {
				return errors.New("output file required (-o)")
			}

			glob, _ := cmd.Flags().GetString("glob")
			if glob == "" {
				glob = "*"
			}

			globex := regexp.MustCompile(glob)

			if len(args) == 0 {
				panic("????")
			}

			key, _ := cmd.Flags().GetString("key")
			rpa := arpy.New(s2i64(key))

			log.Printf("building %s: key=%x, glob=%s", out, key, glob)

			for i := range args {
				log.Printf("walking %s", args[i])

				err := filepath.Walk(args[i], func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if info.IsDir() {
						return nil
					}

					if !globex.MatchString(path) {
						return nil
					}

					// include this file
					fp, err := os.Open(path)
					if err != nil {
						return err
					}

					data, err := ioutil.ReadAll(fp)
					if err != nil {
						return err
					}

					log.Printf("include: %s", path)
					rpa.AddFile(path, data)

					return nil
				})

				if err != nil {
					return fmt.Errorf("error discovering and adding files: %s", err)
				}
			}

			log.Printf("writing...")
			outfp, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("could not write to %s: %s", out, err)
			}

			rpa.WriteTo(outfp)
			outfp.Sync()
			outfp.Close()
			log.Printf("done")

			return nil
		},
	}

	unpack = &cobra.Command{
		Use:   "unpack",
		Short: "unpack an RPA",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, _ := cmd.Flags().GetString("in")
			if in == "" {
				return errors.New("input file required")
			}

			out, _ := cmd.Flags().GetString("out")
			if out == "" {
				out = "/tmp"
			}

			fp, err := os.Open(in)
			if err != nil {
				return fmt.Errorf("could not open %s: %s", in, err)
			}

			rpa, err := arpy.Decode(fp)
			if err != nil {
				return fmt.Errorf("could not parse %s: %s", in, err)
			}

			for i := range rpa.Indexes {
				idx := rpa.Indexes[i]
				dst := filepath.Join(out, idx.Name)

				log.Printf("%s -> %s", idx.Name, dst)
				dfp, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("could not write %s to %s", idx.Name, dst)
				}

				content, err := rpa.FileAt(idx)
				if err != nil {
					return fmt.Errorf("could not get content for %s (%d,%d), bad rpa or bug: %s", idx.Name, idx.Offset, idx.Size, err)
				}

				dfp.Write(content)
				dfp.Sync()
				dfp.Close()
			}

			fp.Close()

			return nil
		},
	}
)

func s2i64(s string) int64 {
	v, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		panic(err)
	}

	return v
}

func init() {
	root.PersistentFlags().StringP("key", "k", "deadbeef", "key for packing or unpacking - expect hex -> int")

	root.AddCommand(pack)
	pack.PersistentFlags().StringP("out", "o", "", "RPA to create")
	pack.PersistentFlags().StringP("glob", "g", "*", "glob pattern to include in archive")

	root.AddCommand(unpack)
	unpack.PersistentFlags().StringP("out", "o", "/tmp", "directory to write files to, defaults to /tmp")
	unpack.PersistentFlags().StringP("in", "i", "", "RPA to extract")
}
