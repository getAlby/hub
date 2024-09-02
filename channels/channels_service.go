package channels

import (
	"context"
	"slices"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"gorm.io/gorm"
)

type channelsService struct {
	db *gorm.DB
}

type ChannelsService interface {
	// TODO: consume channel events
	ListChannels(ctx context.Context, lnClient lnclient.LNClient) (channels []Channel, err error)
}

type Channel struct {
	Open                                     bool
	ChannelSizeSat                           uint64
	LocalBalance                             uint64
	LocalSpendableBalance                    uint64
	RemoteBalance                            uint64
	Id                                       string
	RemotePubkey                             string
	FundingTxId                              string
	Active                                   bool
	Public                                   bool
	InternalChannel                          interface{}
	Confirmations                            *uint32
	ConfirmationsRequired                    *uint32
	ForwardingFeeBaseMsat                    uint32
	UnspendablePunishmentReserve             uint64
	CounterpartyUnspendablePunishmentReserve uint64
	Error                                    *string
	Status                                   string
	IsOutbound                               bool
}

func NewChannelsService(db *gorm.DB) *channelsService {
	return &channelsService{
		db: db,
	}
}

func (svc *channelsService) ListChannels(ctx context.Context, lnClient lnclient.LNClient) (channels []Channel, err error) {

	lnClientChannels, err := lnClient.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	channels = []Channel{}
	for _, lnClientChannel := range lnClientChannels {
		status := "offline"
		if lnClientChannel.Active {
			status = "online"
		} else if lnClientChannel.Confirmations != nil && lnClientChannel.ConfirmationsRequired != nil && *lnClientChannel.ConfirmationsRequired > *lnClientChannel.Confirmations {
			status = "opening"
		}

		// create or update channel in our DB
		var dbChannel db.Channel
		result := svc.db.Limit(1).Find(&dbChannel, &db.Channel{
			ChannelID: lnClientChannel.Id,
		})

		if result.Error != nil {
			return nil, result.Error
		}

		if result.RowsAffected == 0 {
			// channel not saved yet
			dbChannel.ChannelID = lnClientChannel.Id
			dbChannel.PeerID = lnClientChannel.RemotePubkey
			dbChannel.ChannelSizeSat = uint64((lnClientChannel.LocalBalanceMsat + lnClientChannel.RemoteBalanceMsat) / 1000)
			dbChannel.FundingTxID = lnClientChannel.FundingTxId
			dbChannel.Open = true
			svc.db.Create(&dbChannel)
		}

		// update channel with latest status
		svc.db.Model(&dbChannel).Updates(&db.Channel{
			Status: status,
		})

		channels = append(channels, Channel{
			Open:                                     true,
			ChannelSizeSat:                           dbChannel.ChannelSizeSat,
			LocalBalance:                             lnClientChannel.LocalBalanceMsat,
			LocalSpendableBalance:                    lnClientChannel.LocalSpendableBalanceMsat,
			RemoteBalance:                            lnClientChannel.RemoteBalanceMsat,
			Id:                                       lnClientChannel.Id,
			RemotePubkey:                             lnClientChannel.RemotePubkey,
			FundingTxId:                              lnClientChannel.FundingTxId,
			Active:                                   lnClientChannel.Active,
			Public:                                   lnClientChannel.Public,
			InternalChannel:                          lnClientChannel.InternalChannel,
			Confirmations:                            lnClientChannel.Confirmations,
			ConfirmationsRequired:                    lnClientChannel.ConfirmationsRequired,
			ForwardingFeeBaseMsat:                    lnClientChannel.ForwardingFeeBaseMsat,
			UnspendablePunishmentReserve:             lnClientChannel.UnspendablePunishmentReserve,
			CounterpartyUnspendablePunishmentReserve: lnClientChannel.CounterpartyUnspendablePunishmentReserve,
			Error:                                    lnClientChannel.Error,
			IsOutbound:                               lnClientChannel.IsOutbound,
			Status:                                   status,
		})
	}

	// review channels in our db
	dbChannels := []db.Channel{}
	result := svc.db.Find(&dbChannels)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, dbChannel := range dbChannels {
		if !slices.ContainsFunc(channels, func(channel Channel) bool { return channel.Id == dbChannel.ChannelID }) {
			if dbChannel.Open {
				// ideally this should never happen as we should subscribe to events
				// but currently we do not
				result := svc.db.Model(&dbChannel).Updates(&db.Channel{
					Status: "closing",
					Open:   false,
				})
				if result.Error != nil {
					return nil, result.Error
				}
			}

			// show channels closed within the last 2 weeks
			// NOTE: we do not know how much balance the channel had at closing.
			if time.Since(dbChannel.UpdatedAt) < 2*7*24*time.Hour {
				channels = append(channels, Channel{
					ChannelSizeSat:        dbChannel.ChannelSizeSat,
					LocalBalance:          0,
					LocalSpendableBalance: 0,
					RemoteBalance:         0,
					Id:                    dbChannel.ChannelID,
					RemotePubkey:          dbChannel.PeerID,
					FundingTxId:           dbChannel.FundingTxID,
					Active:                false,
					Status:                dbChannel.Status,
				})
			}
		}
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return channels, nil
}
