package functions

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/dbschenker/heap-dump-management/heap-dump-companion/internal/decrypt"
	"github.com/dbschenker/heap-dump-management/heap-dump-companion/internal/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var heapDumpLocation string
var output string
var aesKeyLocation string
var topic string
var transitMountPoint string

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a provided file with a encrypted key using hashicorp vault",
	Long: `Companion implementation intended to work with the general heap dump service.

This command takes a encrypted heap dump, the encrypted AES Key of the heap dump and decrypts both
using the transit engine of hashicorp Vault. 

Examples:

heap-dump-companion decrypt --input-file test/test.dump.crypted --output-file test/test.dump --key test/test.key -t some-tenant`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := vault.GenerateTransitVaultClient()
		cobra.CheckErr(err)
		fullAesKeyLocation, err := filepath.Abs(aesKeyLocation)
		cobra.CheckErr(err)
		encryptedKey, err := os.ReadFile(fullAesKeyLocation)
		cobra.CheckErr(err)
		plainTextKey, err := vault.TransitDecryptString(client, transitMountPoint, viper.GetString("topic"), string(encryptedKey))
		cobra.CheckErr(err)
		decodedKey, err := base64.StdEncoding.DecodeString(plainTextKey)
		cobra.CheckErr(err)
		fullOutputLocation, err := filepath.Abs(output)
		cobra.CheckErr(err)
		fullHeapDumpLocation, err := filepath.Abs(heapDumpLocation)
		cobra.CheckErr(err)
		dir, file := filepath.Split(fullHeapDumpLocation)
		err = decrypt.DecryptFile(os.DirFS(dir), decodedKey, file, fullOutputLocation)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.PersistentFlags().StringVarP(&heapDumpLocation, "input-file", "i", "", "Path to the encrypted heap dump")
	decryptCmd.PersistentFlags().StringVarP(&output, "output-file", "o", "", "Desired output file after decryption")
	decryptCmd.PersistentFlags().StringVarP(&aesKeyLocation, "key", "k", "", "Path to the encrypted key that should be used for dectyption")
	decryptCmd.PersistentFlags().StringVarP(&topic, "topic", "t", "", "Topic/Tenant owner of the heap dump to be decrypted")
	decryptCmd.PersistentFlags().StringVarP(&transitMountPoint, "transit-mount-point", "T", "eaas-heap-dump-service", "Transit engine mount point in vault")

	decryptCmd.MarkFlagRequired("input-file")
	decryptCmd.MarkFlagRequired("output-file")
	decryptCmd.MarkFlagRequired("key")

	viper.BindPFlag("topic", decryptCmd.PersistentFlags().Lookup("topic"))
}
