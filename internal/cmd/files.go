package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
}

var filesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		contentType, _ := cmd.Flags().GetString("content-type")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.FilesListOptions{
			Page:        page,
			PageSize:    pageSize,
			ContentType: contentType,
			Status:      status,
		}

		resp, err := client.ListFiles(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "FILENAME", "MIME TYPE", "SIZE", "STATUS", "CREATED"}
		var rows [][]string
		for _, f := range resp.Items {
			size := formatFileSize(f.FileSize)
			rows = append(rows, []string{
				f.ID,
				f.Filename,
				f.MimeType,
				size,
				string(f.Status),
				f.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d files\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var filesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get file details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		file, err := client.GetFile(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get file: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(file)
		}

		fmt.Printf("File ID:      %s\n", file.ID)
		fmt.Printf("Filename:     %s\n", file.Filename)
		fmt.Printf("MIME Type:    %s\n", file.MimeType)
		fmt.Printf("File Size:    %s\n", formatFileSize(file.FileSize))
		fmt.Printf("URL:          %s\n", file.URL)
		fmt.Printf("Alt:          %s\n", file.Alt)
		fmt.Printf("Status:       %s\n", file.Status)
		if file.ContentType != "" {
			fmt.Printf("Content Type: %s\n", file.ContentType)
		}
		if file.Width > 0 && file.Height > 0 {
			fmt.Printf("Dimensions:   %dx%d\n", file.Width, file.Height)
		}
		fmt.Printf("Created:      %s\n", file.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", file.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var filesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, _ := cmd.Flags().GetString("filename")
		url, _ := cmd.Flags().GetString("url")
		alt, _ := cmd.Flags().GetString("alt")
		contentType, _ := cmd.Flags().GetString("content-type")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create file '%s'\n", filename)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.FileCreateRequest{
			Filename:    filename,
			URL:         url,
			Alt:         alt,
			ContentType: contentType,
		}

		file, err := client.CreateFile(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(file)
		}

		fmt.Printf("Created file %s\n", file.ID)
		fmt.Printf("Filename: %s\n", file.Filename)
		fmt.Printf("Status:   %s\n", file.Status)

		return nil
	},
}

var filesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete file %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete file %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteFile(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		fmt.Printf("Deleted file %s\n", args[0])
		return nil
	},
}

// formatFileSize formats a file size in bytes to a human-readable string.
func formatFileSize(size int64) string {
	if size == 0 {
		return "0 B"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(filesCmd)

	filesCmd.AddCommand(filesListCmd)
	filesListCmd.Flags().String("content-type", "", "Filter by content type (image, video, document)")
	filesListCmd.Flags().String("status", "", "Filter by status (pending, ready, failed, processing)")
	filesListCmd.Flags().Int("page", 1, "Page number")
	filesListCmd.Flags().Int("page-size", 20, "Results per page")

	filesCmd.AddCommand(filesGetCmd)

	filesCmd.AddCommand(filesCreateCmd)
	filesCreateCmd.Flags().String("filename", "", "File name (required)")
	filesCreateCmd.Flags().String("url", "", "Source URL for the file")
	filesCreateCmd.Flags().String("alt", "", "Alt text for the file")
	filesCreateCmd.Flags().String("content-type", "", "Content type (image, video, document)")
	_ = filesCreateCmd.MarkFlagRequired("filename")

	filesCmd.AddCommand(filesDeleteCmd)
}
