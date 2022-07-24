package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (warlock *Warlock) CanConflagrate(sim *core.Simulation) bool {
	return warlock.Talents.Conflagrate && warlock.ImmolateDot.IsActive() && warlock.Conflagrate.IsReady(sim)
}

func (warlock *Warlock) registerConflagrateSpell() {

	baseCost := 0.16 * warlock.BaseMana
	costReductionFactor := 1.0
	if float64(warlock.Talents.Cataclysm) > 0 {
		costReductionFactor -= 0.01 + 0.03*float64(warlock.Talents.Cataclysm)
	}
	spellCoefficient:= 0.2 + 0.04*float64(warlock.Talents.ShadowAndFlame)

	actionID := core.ActionID{SpellID: 17962}
	target := warlock.CurrentTarget

	effect := core.SpellEffect{
		ProcMask:             core.ProcMaskSpellDamage,
		BonusSpellCritRating: 5*(core.TernaryFloat64(warlock.Talents.Devastation, 1, 0) + float64(warlock.Talents.FireAndBrimstone))*core.CritRatingPerCritChance,
		DamageMultiplier: 	  0.6,
		ThreatMultiplier: 	  1 - 0.1*float64(warlock.Talents.DestructiveReach),
		BaseDamage:       	  core.BaseDamageConfigMagicNoRoll(785, spellCoefficient*5),
		OutcomeApplier:   	  warlock.OutcomeFuncMagicHitAndCrit(warlock.SpellCritMultiplier(1, float64(warlock.Talents.Ruin)/5)),
		OnInit: func(sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
			spellEffect.DamageMultiplier = 0.6 * warlock.spellDamageMultiplierHelper(sim, spell, spellEffect)
		},
		OnSpellHitDealt:  applyDotOnLanded(&warlock.ConflagrateDot),
	}

	warlock.Conflagrate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:     actionID,
		SpellSchool:  core.SpellSchoolFire,
		ResourceType: stats.Mana,
		BaseCost:     baseCost,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: baseCost * costReductionFactor,
				GCD:  core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Second * 10,
			},
			OnCastComplete: func(sim *core.Simulation, spell *core.Spell) {
				if !warlock.HasMajorGlyph(proto.WarlockMajorGlyph_GlyphOfConflagrate) {
					warlock.ImmolateDot.Deactivate(sim)
					//warlock.ShadowflameDot.Deactivate(sim)
				}
			},
		},

		ApplyEffects: core.ApplyEffectFuncDirectDamage(effect),
	})

	warlock.ConflagrateDot = core.NewDot(core.Dot{
		Spell: warlock.Conflagrate,
		Aura: target.RegisterAura(core.Aura{
			Label:    "conflagrate-" + strconv.Itoa(int(warlock.Index)),
			ActionID: actionID,
		}),
		NumberOfTicks: 3,
		TickLength:    time.Second * 2,
		TickEffects: core.TickFuncSnapshot(target, core.SpellEffect{
			DamageMultiplier: 0.4/3,
			ThreatMultiplier: 1 - 0.1*float64(warlock.Talents.DestructiveReach),
			BaseDamage:       core.BaseDamageConfigMagicNoRoll(785, spellCoefficient*5),
			OutcomeApplier:   warlock.OutcomeFuncTick(),
			IsPeriodic:       true,
			ProcMask:         core.ProcMaskPeriodicDamage,
			OnInit: func(sim *core.Simulation, spell *core.Spell, spellEffect *core.SpellEffect) {
				spellEffect.DamageMultiplier = 0.4/3 * warlock.spellDamageMultiplierHelper(sim, spell, spellEffect)
			},
		}),
	})
}
