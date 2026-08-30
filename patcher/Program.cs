using Mono.Cecil;
using Mono.Cecil.Cil;
using System;
using System.IO;
using System.Linq;

var dll = "/data/games/mathbreakers/Mathbreakers_Data/Managed/Assembly-CSharp.dll";
var managedDir = Path.GetDirectoryName(dll);

var resolver = new DefaultAssemblyResolver();
resolver.AddSearchDirectory(managedDir);

var asm = AssemblyDefinition.ReadAssembly(dll, new ReaderParameters {
    ReadWrite = true,
    AssemblyResolver = resolver
});

var type = asm.MainModule.Types.First(t => t.Name == "EndLevelTrigger");
var method = type.Methods.First(m => m.Name == "OnTriggerEnter");
var il = method.Body.GetILProcessor();

// Find the Application.LoadLevel call
var loadLevel = method.Body.Instructions
    .First(i => i.OpCode == OpCodes.Call && i.Operand.ToString().Contains("LoadLevel"));

// 1. Resolve mscorlib directly from the game's local dependencies
var mscorlibRef = asm.MainModule.AssemblyReferences.First(r => r.Name == "mscorlib");
var mscorlibAsm = resolver.Resolve(mscorlibRef);
var fileType = mscorlibAsm.MainModule.GetType("System.IO.File");

// 2. Safely get the WriteAllText(string, string) signature
var writeAllTextDef = fileType.Methods.First(m => m.Name == "WriteAllText" && m.Parameters.Count == 2);
var writeAllTextMethod = asm.MainModule.ImportReference(writeAllTextDef);

// 3. Create the instructions using a relative path
// Wine handles relative paths by placing them in the application's launch directory
var ins1 = il.Create(OpCodes.Ldstr, "level_cleared.txt");
var ins2 = il.Create(OpCodes.Ldstr, "triggered");
var ins3 = il.Create(OpCodes.Call, writeAllTextMethod);

// Insert right before LoadLevel
il.InsertBefore(loadLevel, ins1);
il.InsertBefore(loadLevel, ins2);
il.InsertBefore(loadLevel, ins3);

asm.Write();
Console.WriteLine("Patcher successfully swapped to Relative File Writer!");