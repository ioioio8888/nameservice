package nameservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle a message to set name
func handleMsgSetName(ctx sdk.Context, keeper Keeper, msg MsgSetName) sdk.Result {
	if !msg.Owner.Equals(keeper.GetOwner(ctx, msg.Name)) { // Checks if the the msg sender is the same as the current owner
		return sdk.ErrUnauthorized("Incorrect Owner").Result() // If not, throw an error
	}
	keeper.SetName(ctx, msg.Name, msg.Value) // If so, set the name to the value specified in the msg.
	return sdk.Result{}                      // return
}

// MsgBuyName defines the BuyName message
type MsgBuyName struct {
	Name  string         `json:"name"`
	Bid   sdk.Coins      `json:"bid"`
	Buyer sdk.AccAddress `json:"buyer"`
}

// NewMsgBuyName is the constructor function for MsgBuyName
func NewMsgBuyName(name string, bid sdk.Coins, buyer sdk.AccAddress) MsgBuyName {
	return MsgBuyName{
		Name:  name,
		Bid:   bid,
		Buyer: buyer,
	}
}

// Route should return the name of the module
func (msg MsgBuyName) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBuyName) Type() string { return "buy_name" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBuyName) ValidateBasic() sdk.Error {
	if msg.Buyer.Empty() {
		return sdk.ErrInvalidAddress(msg.Buyer.String())
	}
	if len(msg.Name) == 0 {
		return sdk.ErrUnknownRequest("Name cannot be empty")
	}
	if !msg.Bid.IsAllPositive() {
		return sdk.ErrInsufficientCoins("Bids must be positive")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBuyName) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgBuyName) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Buyer}
}

// Handle a message to buy name
func handleMsgBuyName(ctx sdk.Context, keeper Keeper, msg MsgBuyName) sdk.Result {
	if keeper.GetPrice(ctx, msg.Name).IsAllGT(msg.Bid) { // Checks if the the bid price is greater than the price paid by the current owner
		return sdk.ErrInsufficientCoins("Bid not high enough").Result() // If not, throw an error
	}
	if keeper.HasOwner(ctx, msg.Name) {
		err := keeper.CoinKeeper.SendCoins(ctx, msg.Buyer, keeper.GetOwner(ctx, msg.Name), msg.Bid)
		if err != nil {
			return sdk.ErrInsufficientCoins("Buyer does not have enough coins").Result()
		}
	} else {
		_, err := keeper.CoinKeeper.SubtractCoins(ctx, msg.Buyer, msg.Bid) // If so, deduct the Bid amount from the sender
		if err != nil {
			return sdk.ErrInsufficientCoins("Buyer does not have enough coins").Result()
		}
	}
	keeper.SetOwner(ctx, msg.Name, msg.Buyer)
	keeper.SetPrice(ctx, msg.Name, msg.Bid)
	return sdk.Result{}
}

// NewHandler returns a handler for "nameservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetName:
			return handleMsgSetName(ctx, keeper, msg)
		case MsgBuyName:
			return handleMsgBuyName(ctx, keeper, msg)
		case MsgDeleteName:
			return handleMsgDeleteName(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to delete name
func handleMsgDeleteName(ctx sdk.Context, keeper Keeper, msg MsgDeleteName) sdk.Result {
	if !keeper.IsNamePresent(ctx, msg.Name) {
		return types.ErrNameDoesNotExist(types.DefaultCodespace).Result()
	}
	if !msg.Owner.Equals(keeper.GetOwner(ctx, msg.Name)) { 
		return sdk.ErrUnauthorized("Incorrect Owner").Result() 
	}
	keeper.DeleteWhois(ctx, msg.Name) 
	return sdk.Result{}
}