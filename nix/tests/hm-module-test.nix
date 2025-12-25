# home manager module tests
#
# run with: nix build .#checks.x86_64-linux.hm-module
#
# tests cover:
# - module evaluation with various configurations
# - type validation for all options
# - generated TOML content verification
# - edge cases (disabled, empty settings, full settings)
{ pkgs, lib, hmModule }:

let
  # minimal home-manager stub for standalone module testing
  hmLib = {
    hm.types.dagOf = lib.types.attrsOf;
    dag.entryAnywhere = x: x;
  };

  # evaluate the hm module with a given config
  evalModule = testConfig:
    lib.evalModules {
      modules = [
        hmModule
        # stub home-manager's expected structure
        {
          options = {
            home.packages = lib.mkOption {
              type = lib.types.listOf lib.types.package;
              default = [ ];
            };
            xdg.configFile = lib.mkOption {
              type = lib.types.attrsOf (lib.types.submodule {
                options = {
                  source = lib.mkOption { type = lib.types.path; };
                  text = lib.mkOption { type = lib.types.str; default = ""; };
                };
              });
              default = { };
            };
          };
        }
        testConfig
      ];
      specialArgs = { inherit pkgs lib; };
    };

  # read generated TOML file content
  readGeneratedToml = evalResult:
    let
      configFile = evalResult.config.xdg.configFile."snitch/snitch.toml" or null;
    in
    if configFile != null && configFile ? source
    then builtins.readFile configFile.source
    else null;

  # test cases
  tests = {
    # test 1: module evaluates when disabled
    moduleDisabled = {
      name = "module-disabled";
      config = {
        programs.snitch.enable = false;
      };
      assertions = evalResult: [
        {
          assertion = evalResult.config.home.packages == [ ];
          message = "packages should be empty when disabled";
        }
        {
          assertion = !(evalResult.config.xdg.configFile ? "snitch/snitch.toml");
          message = "config file should not exist when disabled";
        }
      ];
    };

    # test 2: module evaluates with enable only (defaults)
    moduleEnabledDefaults = {
      name = "module-enabled-defaults";
      config = {
        programs.snitch.enable = true;
      };
      assertions = evalResult: [
        {
          assertion = builtins.length evalResult.config.home.packages == 1;
          message = "package should be installed when enabled";
        }
      ];
    };

    # test 3: all theme values are valid
    themeValidation = {
      name = "theme-validation";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults.theme = "catppuccin-mocha";
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = toml != null;
            message = "TOML config should be generated";
          }
          {
            assertion = lib.hasInfix "catppuccin-mocha" toml;
            message = "theme should be set in TOML";
          }
        ];
    };

    # test 4: full configuration with all options
    fullConfiguration = {
      name = "full-configuration";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults = {
            interval = "2s";
            numeric = true;
            fields = [ "pid" "process" "proto" ];
            theme = "nord";
            units = "si";
            color = "always";
            resolve = false;
            dns_cache = false;
            ipv4 = true;
            ipv6 = false;
            no_headers = true;
            output_format = "json";
            sort_by = "pid";
          };
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = toml != null;
            message = "TOML config should be generated";
          }
          {
            assertion = lib.hasInfix "interval = \"2s\"" toml;
            message = "interval should be 2s";
          }
          {
            assertion = lib.hasInfix "numeric = true" toml;
            message = "numeric should be true";
          }
          {
            assertion = lib.hasInfix "theme = \"nord\"" toml;
            message = "theme should be nord";
          }
          {
            assertion = lib.hasInfix "units = \"si\"" toml;
            message = "units should be si";
          }
          {
            assertion = lib.hasInfix "color = \"always\"" toml;
            message = "color should be always";
          }
          {
            assertion = lib.hasInfix "resolve = false" toml;
            message = "resolve should be false";
          }
          {
            assertion = lib.hasInfix "output_format = \"json\"" toml;
            message = "output_format should be json";
          }
          {
            assertion = lib.hasInfix "sort_by = \"pid\"" toml;
            message = "sort_by should be pid";
          }
        ];
    };

    # test 5: output format enum validation
    outputFormatCsv = {
      name = "output-format-csv";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults.output_format = "csv";
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = lib.hasInfix "output_format = \"csv\"" toml;
            message = "output_format should accept csv";
          }
        ];
    };

    # test 6: units enum validation
    unitsIec = {
      name = "units-iec";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults.units = "iec";
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = lib.hasInfix "units = \"iec\"" toml;
            message = "units should accept iec";
          }
        ];
    };

    # test 7: color never value
    colorNever = {
      name = "color-never";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults.color = "never";
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = lib.hasInfix "color = \"never\"" toml;
            message = "color should accept never";
          }
        ];
    };

    # test 8: freeform type allows custom keys
    freeformCustomKeys = {
      name = "freeform-custom-keys";
      config = {
        programs.snitch = {
          enable = true;
          settings = {
            defaults.theme = "dracula";
            custom_section = {
              custom_key = "custom_value";
            };
          };
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = lib.hasInfix "custom_key" toml;
            message = "freeform type should allow custom keys";
          }
        ];
    };

    # test 9: all themes evaluate correctly
    allThemes =
      let
        themes = [
          "ansi"
          "catppuccin-mocha"
          "catppuccin-macchiato"
          "catppuccin-frappe"
          "catppuccin-latte"
          "gruvbox-dark"
          "gruvbox-light"
          "dracula"
          "nord"
          "tokyo-night"
          "tokyo-night-storm"
          "tokyo-night-light"
          "solarized-dark"
          "solarized-light"
          "one-dark"
          "mono"
          "auto"
        ];
      in
      {
        name = "all-themes";
        # use the last theme as the test config
        config = {
          programs.snitch = {
            enable = true;
            settings.defaults.theme = "auto";
          };
        };
        assertions = evalResult:
          let
            # verify all themes can be set by evaluating them
            themeResults = map
              (theme:
                let
                  result = evalModule {
                    programs.snitch = {
                      enable = true;
                      settings.defaults.theme = theme;
                    };
                  };
                  toml = readGeneratedToml result;
                in
                {
                  inherit theme;
                  success = toml != null && lib.hasInfix theme toml;
                }
              )
              themes;
            allSucceeded = lib.all (r: r.success) themeResults;
          in
          [
            {
              assertion = allSucceeded;
              message = "all themes should evaluate correctly: ${
                lib.concatMapStringsSep ", " 
                  (r: "${r.theme}=${if r.success then "ok" else "fail"}") 
                  themeResults
              }";
            }
          ];
      };

    # test 10: fields list serialization
    fieldsListSerialization = {
      name = "fields-list-serialization";
      config = {
        programs.snitch = {
          enable = true;
          settings.defaults.fields = [ "pid" "process" "proto" "state" ];
        };
      };
      assertions = evalResult:
        let
          toml = readGeneratedToml evalResult;
        in
        [
          {
            assertion = lib.hasInfix "pid" toml && lib.hasInfix "process" toml;
            message = "fields list should be serialized correctly";
          }
        ];
    };
  };

  # run all tests and collect results
  runTests =
    let
      testResults = lib.mapAttrsToList
        (name: test:
          let
            evalResult = evalModule test.config;
            assertions = test.assertions evalResult;
            failures = lib.filter (a: !a.assertion) assertions;
          in
          {
            inherit name;
            testName = test.name;
            passed = failures == [ ];
            failures = map (f: f.message) failures;
          }
        )
        tests;

      allPassed = lib.all (r: r.passed) testResults;
      failedTests = lib.filter (r: !r.passed) testResults;

      summary = ''
        ========================================
        home manager module test results
        ========================================
        total tests: ${toString (builtins.length testResults)}
        passed: ${toString (builtins.length (lib.filter (r: r.passed) testResults))}
        failed: ${toString (builtins.length failedTests)}
        ========================================
        ${lib.concatMapStringsSep "\n" (r: 
          if r.passed 
          then "[yes] ${r.testName}"
          else "[no] ${r.testName}\n  ${lib.concatStringsSep "\n  " r.failures}"
        ) testResults}
        ========================================
      '';
    in
    {
      inherit testResults allPassed failedTests summary;
    };

  results = runTests;

in
pkgs.runCommand "hm-module-test"
{
  passthru = {
    inherit results;
    # expose for debugging
    inherit evalModule tests;
  };
}
  (
    if results.allPassed
    then ''
      echo "${results.summary}"
      echo "all tests passed"
      touch $out
    ''
    else ''
      echo "${results.summary}"
      echo ""
      echo "failed tests:"
      ${lib.concatMapStringsSep "\n" (t: ''
        echo "  - ${t.testName}: ${lib.concatStringsSep ", " t.failures}"
      '') results.failedTests}
      exit 1
    ''
  )

