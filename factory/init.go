package factoy

import (
	"fmt"
	"log"
	dbg "runtime/debug"

	"hero-emulator/auth"
	"hero-emulator/database"
	"hero-emulator/npc"
	"hero-emulator/player"
	"hero-emulator/utils"
)

type Factory interface {
	Handle(*database.Socket, []byte) ([]byte, error)
}

var (
	pkgTypes = map[uint16]Factory{
		000:   &auth.LoginHandler{},
		002:   &auth.ListServersHandler{},
		004:   &auth.SelectServerHandler{},
		257:   &auth.ListCharactersHandler{},
		258:   &auth.CancelCharacterCreationHandler{},
		259:   &auth.CharacterCreationHandler{},
		261:   &auth.CharacterSelectionHandler{},
		434:   &auth.CharacterDeletionHandler{},
		522:   &player.QuestHandler{},
		441:   &player.InTacticalSpaceTPHandler{},
		2310:  &player.QuitGameHandler{},
		2312:  &player.ServerMenuHandler{},
		2313:  &player.CharacterMenuHandler{},
		4609:  &player.RespawnHandler{},
		4612:  &player.RespawnHandler{},
		8705:  &player.MovementHandler{},
		8706:  &player.MovementHandler{},
		9732:  &player.MovementHandler{},
		10257: &player.OpenTacticalSpaceHandler{},
		10753: &player.SendPvPRequestHandler{},
		10754: &player.RespondPvPRequestHandler{},
		15616: &npc.OpenConsignmentHandler{},
		15617: &npc.RegisterItemHandler{},
		15618: &npc.BuyConsignmentItemHandler{},
		15619: &npc.ClaimMenuHandler{},
		15620: &npc.ClaimConsignmentItemHandler{},
		16896: &player.CastSkillHandler{},
		16899: &player.CastSkillHandler{},
		16900: &player.CastSkillHandler{},
		17152: &player.BattleModeHandler{},
		17153: &player.BattleModeHandler{},
		18704: &player.CastMonkSkillHandler{},
		18705: &player.DealDamageHandler{},
		19715: &player.RemoveBuffHandler{},
		20482: &player.TacticalSpaceTPHandler{},
		20737: &player.ToggleMountPetHandler{},
		20738: &player.TogglePetHandler{},
		20741: &player.PetCombatModeHandler{},
		20993: &player.SendPartyRequestHandler{},
		20994: &player.RespondPartyRequestHandler{},
		20995: &player.LeavePartyHandler{},
		20998: &player.ExpelFromPartyHandler{},
		21249: &player.SendTradeRequestHandler{},
		21250: &player.RespondTradeRequestHandler{},
		21251: &player.CancelTradeHandler{},
		21252: &player.AddTradeItemHandler{},
		21254: &player.AddTradeGoldHandler{},
		21256: &player.RemoveTradeItemHandler{},
		21257: &player.AcceptTradeHandler{},
		21506: &npc.StrengthenHandler{},
		21508: &npc.ProductionHandler{},
		21509: &npc.DismantleHandler{},
		21510: &npc.ExtractionHandler{},
		21513: &npc.AdvancedFusionHandler{},
		21520: &player.HolyWaterUpgradeHandler{},
		21526: &npc.CreateSocketHandler{},
		21527: &npc.UpgradeSocketHandler{},
		21536: &npc.CoProductionHandler{},
		21538: &npc.AppearanceHandler{},
		21761: &player.OpenSaleHandler{},
		21762: &player.CloseSaleHandler{},
		21763: &player.VisitSaleHandler{},
		21764: &player.BuySaleItemHandler{},
		21769: &player.OpenSaleMenuHandler{},
		22273: &npc.OpenHandler{},
		22274: &npc.PressButtonHandler{},
		22529: &npc.BuyItemHandler{},
		22530: &npc.SellItemHandler{},
		22785: &player.LootHandler{},
		22786: &player.RemoveItemHandler{},
		22787: &player.ReplaceItemHandler{},
		22788: &player.UseConsumableHandler{},
		22789: &player.SwitchWeaponHandler{},
		22790: &player.CombineItemsHandler{},
		22791: &player.SwapItemsHandler{},
		22792: &player.OpenBoxHandler2{},
		22793: &player.SplitItemHandler{},
		22800: &player.OpenBoxHandler{},
		22801: &player.DressUpHandler{},
		22806: &player.ActivateTimeLimitedItemHandler{},
		22817: &player.ActivateTimeLimitedItemHandler2{},
		22832: &player.DestroyItemHandler{},
		22848: &player.ReplaceHTItemHandler{},
		22849: &player.DiscriminateItemHandler{},
		25089: &player.InspectItemHandler{},
		25090: &auth.StartGameHandler{},
		25345: &player.DepositHandler{},
		25346: &player.WithdrawHandler{},
		25601: &player.OpenHTMenuHandler{},
		25602: &player.CloseHTMenuHandler{},
		25604: &player.BuyHTItemHandler{},
		26633: &player.OpenBuyMenuHandler{},
		28929: &player.ChatHandler{},
		28930: &player.ChatHandler{},
		28931: &player.ChatHandler{},
		28932: &player.ChatHandler{},
		28933: &player.ChatHandler{},
		28935: &player.ChatHandler{},
		28943: &player.ChatHandler{},
		28945: &player.ChatHandler{},
		28946: &player.ChatHandler{},
		30721: &player.ArrangeFunctionHandler{},
		32769: &player.ArrangeBankHandler{},
		33026: &player.UpgradeSkillHandler{},
		33027: &player.DowngradeSkillHandler{},
		33029: &player.DivineUpgradeSkillHandler{},
		33030: &player.RemoveSkillHandler{},
		33282: &player.UpgradePassiveSkillHandler{},
		33283: &player.DowngradePassiveSkillHandler{},
		33284: &player.RemovePassiveSkillHandler{},
		33285: &player.MeditationHandler{},
		33537: &player.CreateGuildHandler{},
		33539: &player.GuildRequestHandler{},
		33540: &player.RespondGuildRequestHandler{},
		33542: &player.LeaveGuildHandler{},
		33543: &player.ExpelFromGuildHandler{},
		33547: &player.ChangeGuildLogoHandler{},
		33585: &player.ChangeRoleHandler{},
		47874: &player.TravelToFiveClanArea{},
		41472: &player.OpenLotHandler{},
		41473: &player.OpenLotHandler{},
		42241: &player.TransferSoulHandler{},
		47875: &player.TravelToCastleHandler{},
		50176: &player.AddStatHandler{},
		50177: &player.AddStatHandler{},
		50178: &player.AddStatHandler{},
		50179: &player.AddNatureHandler{},
		50180: &player.AddNatureHandler{},
		50181: &player.AddNatureHandler{},
		52224: &player.SaveSlotbarHandler{},
	}

	pkgTypes2 = map[byte]Factory{
		40:  &player.EnterGateHandler{},
		65:  &player.AttackHandler{},
		68:  &player.InstantAttackHandler{},
		69:  &player.AttackHandler{},
		207: &player.TargetSelectionHandler{},
		250: &player.AidHandler{},
		254: &player.DealDamageHandler{},
	}
)

func init() {

	database.Handler = func(s *database.Socket, data []byte, pkgType uint16) ([]byte, error) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println()
				log.Println(err)
				log.Printf("%+v", string(dbg.Stack()))

				if s.User != nil {
					log.Printf("User ID: %s", s.User.ID)
				}
				if s.Character != nil {
					log.Printf("Character ID: %d\t Character Name: %s", s.Character.ID, s.Character.Name)
				}

				fmt.Printf("Data: ")
				r := utils.Packet{}
				r.Concat(data)
				r.Print()
				fmt.Println()
			}
		}()

		pkg, ok := pkgTypes[pkgType]
		if ok {
			return pkg.Handle(s, data)
		}

		pkgType2 := byte(pkgType / 256)
		pkg, ok = pkgTypes2[pkgType2]
		if ok {
			return pkg.Handle(s, data)
		}

		return nil, nil
	}
}
