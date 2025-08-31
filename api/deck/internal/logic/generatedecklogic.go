package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"
	"github.com/joeblew999/infra/pkg/api/deck/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateDeckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateDeckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateDeckLogic {
	return &GenerateDeckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateDeckLogic) GenerateDeck(req *types.GenerateDeckRequest) (resp *types.GenerateDeckResponse, err error) {
	// Generate unique ID for this deck request
	deckID := fmt.Sprintf("deck-%d", time.Now().UnixNano())
	
	// Generate .dsh content based on description
	dshContent, err := l.generateDshFromDescription(req.Description, req.Width, req.Height, req.Style)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dsh content: %w", err)
	}
	
	// Create response
	resp = &types.GenerateDeckResponse{
		Id:         deckID,
		Status:     "completed", // For now, synchronous processing
		Message:    "Deck generated successfully",
		DshContent: dshContent,
		OutputUrl:  fmt.Sprintf("/api/v1/deck/download/%s?format=%s", deckID, req.Format),
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
	
	// TODO: Store generated content for later download
	// TODO: Process dsh -> xml -> svg/png/pdf using pkg/deck pipeline
	
	return resp, nil
}

// generateDshFromDescription converts a text description into .dsh markup
func (l *GenerateDeckLogic) generateDshFromDescription(description string, width, height int, style string) (string, error) {
	// Simple template-based generation for now
	// In a real implementation, this could use AI/LLM to generate more sophisticated .dsh
	
	dsh := fmt.Sprintf(`deck %d %d
	// Generated from: %s
	// Style: %s
	
	text "Generated Deck" %d %d 3
	text "%s" %d %d 1.5
	
	// Add some basic elements based on description
	`, width, height, description, style, width/2, height-1, description, width/2, height/2)
	
	// Add content based on keywords in description
	if strings.Contains(strings.ToLower(description), "card") || strings.Contains(strings.ToLower(description), "playing") {
		dsh += l.generatePlayingCardElements(width, height)
	}
	
	dsh += "edeck\n"
	return dsh, nil
}

// generatePlayingCardElements adds playing card specific elements to the .dsh
func (l *GenerateDeckLogic) generatePlayingCardElements(width, height int) string {
	return fmt.Sprintf(`
	// Playing card elements
	rect 1 1 %d %d "white" 1
	circle %d %d 0.5 "red" 1
	text "♠ ♥ ♦ ♣" %d %d 2
	`, width-2, height-2, width/4, height/4, width/2, height/3)
}
