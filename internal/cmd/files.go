package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("file", f.ID),
				f.Filename,
				f.MimeType,
				size,
				string(f.Status),
				f.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d files\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "File ID:      %s\n", file.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Filename:     %s\n", file.Filename)
		_, _ = fmt.Fprintf(outWriter(cmd), "MIME Type:    %s\n", file.MimeType)
		_, _ = fmt.Fprintf(outWriter(cmd), "File Size:    %s\n", formatFileSize(file.FileSize))
		_, _ = fmt.Fprintf(outWriter(cmd), "URL:          %s\n", file.URL)
		_, _ = fmt.Fprintf(outWriter(cmd), "Alt:          %s\n", file.Alt)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:       %s\n", file.Status)
		if file.ContentType != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Content Type: %s\n", file.ContentType)
		}
		if file.Width > 0 && file.Height > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Dimensions:   %dx%d\n", file.Width, file.Height)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", file.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", file.UpdatedAt.Format(time.RFC3339))

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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create file '%s'", filename)) {
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created file %s\n", file.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Filename: %s\n", file.Filename)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:   %s\n", file.Status)

		return nil
	},
}

var filesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update file %s", args[0])) {
			return nil
		}

		var req api.FileUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		file, err := client.UpdateFile(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update file: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(file)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated file %s\n", file.ID)
		return nil
	},
}

var filesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete file %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete file %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteFile(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted file %s\n", args[0])
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

	filesCmd.AddCommand(filesUpdateCmd)
	addJSONBodyFlags(filesUpdateCmd)

	filesCmd.AddCommand(filesDeleteCmd)
}
